// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// WebRTC connection and processing module

package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/teonet-go/teogw"
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

	// Connect to teonet peer, send request, get answer and resend answer to
	// tru sender
	proxyRequest := func(dc *teowebrtc_client.DataChannel, gw *teogw.TeogwData) {

		var err error

		// Send answer before return
		defer func() {
			if err != nil {
				gw.SetError(err)
			}
			data, err := json.Marshal(gw)
			if err != nil {
				// TODO: send this error to dc
				return
			}
			dc.Send(data)
		}()

		// Send api request to teonet peer
		data, err := teo.proxyCall(gw.Address, gw.Command, gw.Data)
		if err != nil {
			return
		}

		gw.SetData(data)
	}

	connected := func(peer string, dc *teowebrtc_client.DataChannel) {
		log.Println("connected to", peer)

		dc.OnOpen(func() {
			log.Println("data channel opened", peer)
		})

		// Register text message handling
		dc.OnMessage(func(data []byte) {
			log.Printf("got Message from peer '%s': '%s'\n", peer, string(data))

			// Unmarshal proxy command
			var request teogw.TeogwData
			err := json.Unmarshal(data, &request)
			log.Println("got Message from peer, request:", request, err)
			if err == nil && len(request.Address) > 0 && len(request.Command) > 0 {
				go proxyRequest(dc, &request)
				return
			}

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
