// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WebRTC connection and processing module

package main

import (
	"github.com/teonet-go/teogw"
	"github.com/teonet-go/teowebrtc_server"
)

// WebRTC data and methods receiver
type WebRTC struct {
	*teowebrtc_server.WebRTC
	*Teonet
}

// Connect and start WebRTC proxy
func newWebRTC(teo *Teonet) (w *WebRTC, err error) {

	const name = "server-1"

	// Create WebRTC object
	w = new(WebRTC)
	w.WebRTC, err = teowebrtc_server.New(
		params.signalAddr,
		params.signalAddrTls,
		name,
		new(teogw.TeogwData).MarshalJson,
		new(teogw.TeogwData).UnmarshalJson,
	)
	w.Teonet = teo
	w.ProxyCall = w.Teonet.proxyCall

	// Start and process signal server
	// go teowebrtc_signal.New(params.signalAddr, params.signalAddrTls)
	// time.Sleep(1 * time.Millisecond) // Wait while ws server start

	// Start and process webrtc server
	// err = teowebrtc_server.Connect(params.signalAddr, name, w.connected)
	// if err != nil {
	// 	log.Fatalln("connect error:", err)
	// }

	return
}

// // This WebRTC server default commands
// const (
// 	cmdSubscribe = "subscribe"
// 	cmdClients   = "clients"
// 	cmdList      = "list"
// )

// // WebRTC data and methods receiver
// type WebRTC struct {
// 	teowebrtc_server.Peers
// 	*Teonet
// }

// // Connect and start WebRTC proxy
// func newWebRTC(teo *Teonet) (w *WebRTC, err error) {

// 	// Create WebRTC object
// 	w = new(WebRTC)
// 	w.Peers.Init()
// 	w.Subscribe.Init()
// 	w.Teonet = teo

// 	// Start and process signal server
// 	go teowebrtc_signal.New(params.signalAddr, params.signalAddrTls)
// 	time.Sleep(1 * time.Millisecond) // Wait while ws server start

// 	const name = "server-1"

// 	// Start and process webrtc server
// 	err = teowebrtc_server.Connect(params.signalAddr, name, w.connected)
// 	if err != nil {
// 		log.Fatalln("connect error:", err)
// 	}

// 	return
// }

// // Connected calls when a peer connected and Data channel created
// func (w *WebRTC) connected(peer string, dc *teowebrtc_client.DataChannel) {
// 	log.Println("connected to", peer)

// 	dc.OnOpen(func() {
// 		log.Println("data channel opened", peer)
// 		w.Add(peer, dc)
// 	})

// 	dc.OnClose(func() {
// 		log.Println("data channel closed", peer)
// 		w.Del(peer, dc)
// 	})

// 	// Register text message handling
// 	dc.OnMessage(func(data []byte) {
// 		log.Printf("got message from peer '%s': '%s'\n", peer, string(data))

// 		// Unmarshal proxy command
// 		var request teogw.TeogwData
// 		err := json.Unmarshal(data, &request)
// 		switch {
// 		// Send teonet proxy request
// 		case err == nil && len(request.Address) > 0 && len(request.Command) > 0:
// 			log.Println("got proxy request:", request)
// 			go w.proxyRequest(dc, &request)

// 		// Execute request to this server
// 		case err == nil && len(request.Address) == 0 && len(request.Command) > 0:
// 			log.Println("got server request:", request)
// 			go w.serverRequest(peer, dc, &request)

// 		// Send echo answer
// 		default:
// 			d := []byte("Answer to: ")
// 			data = append(d, data...)
// 			dc.Send(data)
// 		}
// 	})
// }

// // Process teonet proxy request: Connect to teonet peer, send request, get
// // answer and resend answer to tru sender
// func (w *WebRTC) proxyRequest(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData) {

// 	var err error

// 	// Send answer before return
// 	defer w.answer(dc, gw, err)

// 	// Send api request to teonet peer
// 	data, err := w.proxyCall(gw.Address, gw.Command, gw.Data)
// 	if err != nil {
// 		return
// 	}

// 	gw.SetData(data)
// }

// // Process this server request
// func (w *WebRTC) serverRequest(peer string, dc *teowebrtc_client.DataChannel,
// 	gw *teogw.TeogwData) {

// 	var err error
// 	var data []byte

// 	// Send answer before return
// 	defer w.answer(dc, gw, err)

// 	// Process request
// 	switch gw.Command {

// 	// Get number of clients
// 	case cmdClients:
// 		l := w.Len()
// 		data = []byte(fmt.Sprintf("%d", l))

// 	// Get list of clients
// 	case cmdList:
// 		data, err = w.getList()

// 	// Subscribe to event
// 	case cmdSubscribe:
// 		w.subscribeRequest(peer, dc, gw)
// 		data = []byte("done")

// 	// Wrong request
// 	default:
// 		err = errors.New("wrong request")
// 	}
// 	if err == nil {
// 		gw.SetData(data)
// 	}
// }

// // getList return json encoded list of clients
// func (w *WebRTC) getList() ([]byte, error) {
// 	type List []string
// 	var list List
// 	for p := range w.ListCh() {
// 		list = append(list, p.Name)
// 	}
// 	return json.Marshal(list)
// }

// // Process this server subscribe request
// func (w *WebRTC) subscribeRequest(peer string, dc *teowebrtc_client.DataChannel,
// 	gw *teogw.TeogwData) {

// 	request := string(gw.Data)
// 	log.Println("got subscribe request:", request)
// 	switch request {
// 	case cmdClients:
// 		w.Onchange(peer, dc, func() {
// 			l := w.Len()
// 			data := []byte(fmt.Sprintf("%d", l))
// 			gw.Command = request
// 			gw.SetData(data)
// 			w.answer(dc, gw, nil)
// 		})
// 	case cmdList:
// 		w.Onchange(peer, dc, func() {
// 			data, err := w.getList()
// 			gw.Command = request
// 			gw.SetData(data)
// 			w.answer(dc, gw, err)
// 		})
// 	}
// }

// // Send answer to data channel
// func (w *WebRTC) answer(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData,
// 	inerr error) (err error) {

// 	if inerr != nil {
// 		gw.SetError(inerr)
// 	}
// 	data, err := json.Marshal(gw)
// 	if inerr != nil {
// 		// TODO: send this error to dc
// 		return
// 	}
// 	err = dc.Send(data)
// 	return
// }
