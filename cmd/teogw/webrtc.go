// Copyright 2022-2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
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

	// Add WebRTC commands
	w.Commands.
		Add("hello",
			func(gw teowebrtc_server.WebRTCData) (data []byte, err error) {
				data = []byte("hello")
				return
			}).
		Add("hello-2",
			func(gw teowebrtc_server.WebRTCData) (data []byte, err error) {
				data = []byte("hello-2")
				return
			})

	return
}
