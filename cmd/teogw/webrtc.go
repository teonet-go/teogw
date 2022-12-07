// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WebRTC connection and processing module

package main

import (
	"log"
	"time"

	"github.com/teonet-go/teowebrtc_client"
	"github.com/teonet-go/teowebrtc_server"
	"github.com/teonet-go/teowebrtc_signal"
)

type WebRTC struct {
}

// Connect and start WebRTC proxy
func newWebRTC(teo *Teonet) (w *WebRTC, err error) {

	// Start and process signal server
	go teowebrtc_signal.New(params.signalAddr, params.signalAddrTls)
	time.Sleep(1 * time.Millisecond) // Wait while ws server start

	const name = "server-1"
	// const url = "localhost:8081"
	connected := func(peer string, dc *teowebrtc_client.DataChannel) {
		log.Println("connected to", peer)

		dc.OnOpen(func() {
			log.Println("data channel opened", peer)
		})

		// Register text message handling
		dc.OnMessage(func(data []byte) {
			log.Printf("got Message from peer '%s': '%s'\n", peer, string(data))
			// Send echo answer
			d := []byte("Answer to: ")
			data = append(d, data...)
			dc.Send(data)
		})
	}

	// Start and process webrtc server
	err = teowebrtc_server.Connect( /* url */ params.signalAddr, name, connected)
	if err != nil {
		log.Fatalln("connect error:", err)
	}

	return
}
