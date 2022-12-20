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

const (
	cmdSubscribe = "subscribe"
	cmdClients   = "clients"
	cmdList      = "list"
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
	w.peers.init()
	w.subscribe.init()
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
		w.del(peer, dc)
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
			go w.serverRequest(peer, dc, &request)

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
func (w *WebRTC) serverRequest(peer string, dc *teowebrtc_client.DataChannel,
	gw *teogw.TeogwData) {

	var err error
	var data []byte

	// Send answer before return
	defer w.answer(dc, gw, err)

	// Process request
	switch gw.Command {

	// Get number of clients
	case cmdClients:
		l := w.len()
		data = []byte(fmt.Sprintf("%d", l))

	// Get list of clients
	case cmdList:
		data, err = w.getList()

	// Subscribe to event
	case cmdSubscribe:
		w.subscribeRequest(peer, dc, gw)
		data = []byte("done")

	// Wrong request
	default:
		err = errors.New("wrong request")
	}
	if err == nil {
		gw.SetData(data)
	}
}

// getList return json encoded list of clients
func (w *WebRTC) getList() ([]byte, error) {
	type List []string
	var list List
	for p := range w.listCh() {
		list = append(list, p.name)
	}
	return json.Marshal(list)
}

// Process this server subscribe request
func (w *WebRTC) subscribeRequest(peer string, dc *teowebrtc_client.DataChannel,
	gw *teogw.TeogwData) {

	request := string(gw.Data)
	log.Println("got subscribe request:", request)
	switch request {
	case cmdClients:
		w.onchange(peer, dc, func() {
			l := w.len()
			data := []byte(fmt.Sprintf("%d", l))
			gw.Command = request
			gw.SetData(data)
			w.answer(dc, gw, nil)
		})
	case cmdList:
		w.onchange(peer, dc, func() {
			data, err := w.getList()
			gw.Command = request
			gw.SetData(data)
			w.answer(dc, gw, err)
		})
	}
}

// Send answer to data channel
func (w *WebRTC) answer(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData,
	inerr error) (err error) {

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

// Init peers object
func (p *peers) init() {
	p.peersMap = make(peersMap)
	p.RWMutex = new(sync.RWMutex)
}

// Add peer to peers map
func (p *peers) add(peer string, dc *teowebrtc_client.DataChannel) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()

	// Close data channel to existing connection from this peer
	if dcCurrent, exists := p.getUnsafe(peer); exists && dcCurrent != dc {
		log.Println("close existing data channel with peer " + peer)
		p.delUnsafe(peer, dcCurrent)
		dcCurrent.Close()
	}

	p.peersMap[peer] = dc
}

// Delete peer from peers map
func (p *peers) del(peer string, dc *teowebrtc_client.DataChannel) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()

	dcCurrent, exists := p.getUnsafe(peer)
	if exists && dcCurrent == dc {
		log.Println("remove peer " + peer)
		p.delUnsafe(peer, dc)
	}
}
func (p *peers) delUnsafe(peer string, dc *teowebrtc_client.DataChannel) {
	delete(p.peersMap, peer)
	go p.subscribe.del(peer, dc)
}

// Get peers dc from map
func (p *peers) getUnsafe(name string) (dc *teowebrtc_client.DataChannel, exists bool) {
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
func (p *peers) onchange(peer string, dc *teowebrtc_client.DataChannel, f func()) {
	log.Println(peer + " subscribed to clients")
	p.subscribe.add(peer, dc, f)
}

// Executes when peers map changed
func (p *peers) changed() {
	for _, sd := range p.subscribe.subscribeMap {
		sd.f()
	}
}

// Subscribe data structure and method receiver
type subscribe struct {
	subscribeID int
	subscribeMap
	*sync.RWMutex
}
type subscribeMap map[int]subscribeData
type subscribeData struct {
	peer string
	dc   *teowebrtc_client.DataChannel
	f    func()
}

// Init subscribe object
func (s *subscribe) init() {
	s.subscribeMap = make(subscribeMap)
	s.RWMutex = new(sync.RWMutex)
}

// Add function to subscribe and return subscribe ID
func (s *subscribe) add(peer string, dc *teowebrtc_client.DataChannel, f func()) int {
	s.Lock()
	defer s.Unlock()
	s.subscribeID++
	s.subscribeMap[s.subscribeID] = subscribeData{peer, dc, f}
	return s.subscribeID
}

// Delete from subscribe by ID or Peer name
func (s *subscribe) del(id interface{}, dc ...*teowebrtc_client.DataChannel) {
	s.Lock()
	defer s.Unlock()
	switch v := id.(type) {
	case int:
		delete(s.subscribeMap, v)
	case string:
		for id, md := range s.subscribeMap {
			if md.peer == v && len(dc) > 0 && md.dc == dc[0] {
				delete(s.subscribeMap, id)
			}
		}
	}
}
