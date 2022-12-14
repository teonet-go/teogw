// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WebRTC connection and processing module

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/teonet-go/teogw"
	"github.com/teonet-go/teowebrtc_client"
	"github.com/teonet-go/teowebrtc_server"
	"github.com/teonet-go/teowebrtc_signal"
)

// WebRTC data and methods receiver
type WebRTC struct {
	peers
	*Teonet
}

// Connect and start WebRTC proxy
func newWebRTC(teo *Teonet) (w *WebRTC, err error) {

	// Create WebRTC object
	w = new(WebRTC)
	w.peersMap = make(peersMap)
	w.RWMutex = new(sync.RWMutex)
	w.Teonet = teo

	// Start and process signal server
	go teowebrtc_signal.New(params.signalAddr, params.signalAddrTls)
	time.Sleep(1 * time.Millisecond) // Wait while ws server start

	const name = "server-1"

	// Start and process webrtc server
	err = teowebrtc_server.Connect(params.signalAddr, name, w.connected)
	if err != nil {
		log.Fatalln("connect error:", err)
	}

	return
}

// Connected calls when a peer connected and Data channel created
func (w *WebRTC) connected(peer string, dc *teowebrtc_client.DataChannel) {
	log.Println("connected to", peer)

	dc.OnOpen(func() {
		log.Println("data channel opened", peer)
		w.add(peer, dc)
	})

	dc.OnClose(func() {
		log.Println("data channel closed", peer)
		w.del(peer)
	})

	// Register text message handling
	dc.OnMessage(func(data []byte) {
		log.Printf("got message from peer '%s': '%s'\n", peer, string(data))

		// Unmarshal proxy command
		var request teogw.TeogwData
		err := json.Unmarshal(data, &request)
		switch {
		// Send teonet proxy request
		case err == nil && len(request.Address) > 0 && len(request.Command) > 0:
			log.Println("got proxy request:", request)
			go w.proxyRequest(dc, &request)

		// Execute request to this server
		case err == nil && len(request.Address) == 0 && len(request.Command) > 0:
			log.Println("got server request:", request)
			go w.serverRequest(dc, &request)

		// Send echo answer
		default:
			d := []byte("Answer to: ")
			data = append(d, data...)
			dc.Send(data)
		}
	})
}

// Process teonet proxy request: Connect to teonet peer, send request, get
// answer and resend answer to tru sender
func (w *WebRTC) proxyRequest(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData) {

	var err error

	// Send answer before return
	defer w.answer(dc, gw, err)

	// Send api request to teonet peer
	data, err := w.proxyCall(gw.Address, gw.Command, gw.Data)
	if err != nil {
		return
	}

	gw.SetData(data)
}

// Process this server request
func (w *WebRTC) serverRequest(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData) {

	var err error
	var data []byte

	// Send answer before return
	defer w.answer(dc, gw, err)

	// Process request
	switch gw.Command {

	// Get number of clients
	case "clients":
		l := w.len()
		data = []byte(fmt.Sprintf("%d", l))

	// Subscribe to event
	case "subscribe":
		w.subscribeRequest(dc, gw)
		data = []byte("done")

	// Wrong request
	default:
		err = errors.New("wrong request")
	}
	if err == nil {
		gw.SetData(data)
	}
}

// Process this server subscribe request
func (w *WebRTC) subscribeRequest(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData) {
	request := string(gw.Data)
	log.Println("got subscribe request:", request)
	switch request {
	case "clients":
		w.onchange(func() {
			l := w.len()
			data := []byte(fmt.Sprintf("%d", l))
			gw.Command = "clients"
			gw.SetData(data)
			w.answer(dc, gw, nil)
		})
	}
}

// Send answer to data channel
func (w *WebRTC) answer(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData, inerr error) (err error) {
	if inerr != nil {
		gw.SetError(inerr)
	}
	data, err := json.Marshal(gw)
	if inerr != nil {
		// TODO: send this error to dc
		return
	}
	err = dc.Send(data)
	return
}

// Peers data and methods receiver
type peers struct {
	peersMap
	*sync.RWMutex
	subscribe
}
type peersMap map[string]*teowebrtc_client.DataChannel
type peerData struct {
	name string
	dc   *teowebrtc_client.DataChannel
}

// Add peer to peers map
func (p *peers) add(peer string, dc *teowebrtc_client.DataChannel) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()
	p.peersMap[peer] = dc
}

// Delete peer from peers map
func (p *peers) del(peer string) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()
	delete(p.peersMap, peer)
}

// Get peers dc from map
func (p *peers) get(name string) (dc *teowebrtc_client.DataChannel, exists bool) {
	p.RLock()
	defer p.RUnlock()
	dc, exists = p.peersMap[name]
	return
}

// Get len of peers map
func (p *peers) len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.peersMap)
}

// Get list channel of peers map
func (p *peers) listCh() (ch chan peerData) {
	p.RLock()
	defer p.RUnlock()
	ch = make(chan peerData)
	go func() {
		for name, dc := range p.peersMap {
			ch <- peerData{name, dc}
		}
		close(ch)
	}()
	return
}

// Get list of peers map
func (p *peers) list() (l []peerData) {
	for p := range p.listCh() {
		l = append(l, p)
	}
	return
}

// Subscribe to change number in peer map
func (p *peers) onchange(f func()) {
	log.Println("subscribed to clients")
	p.subscribe.add(f)
}

// Executes when peers map changed
func (p *peers) changed() {
	for _, f := range p.subscribe.subscribe {
		f()
	}
}

type subscribe struct {
	subscribe []func()
}

func (s *subscribe) add(f func()) {
	s.subscribe = append(s.subscribe, f)
}
