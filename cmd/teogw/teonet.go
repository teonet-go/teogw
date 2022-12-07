// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet connection and processing module

package main

import (
	"fmt"

	"github.com/teonet-go/teomon"
	"github.com/teonet-go/teonet"
)

// Teonet data structure and Methods receiver
type Teonet struct {
	*teonet.Teonet
}

// Connect and start newTeonet
func newTeonet() (teo *Teonet, err error) {

	// Create Teonet connector
	teo = new(Teonet)
	teo.Teonet, err = teonet.New(
		params.appShortName, params.port, teonet.Stat(params.stat), teonet.Hotkey(params.hotkey),
		params.loglevel, teonet.Logfilter(params.logfilter),
	)
	if err != nil {
		return
	}
	fmt.Println("teonet address:", teo.Address())

	// Connect to Teonet
	err = teo.Connect()
	if err != nil {
		fmt.Println("can't connect to Teonet, error:", err)
		return
	}
	fmt.Println("connected to teonet")

	// Connect to monitor if it set in parameters
	if len(monitor) > 0 {
		teomon.Connect(teo, monitor, teomon.Metric{
			AppName:      appName,
			AppShort:     appShort,
			AppVersion:   appVersion,
			TeoVersion:   teonet.Version,
			AppStartTime: appStartTime,
		})
		fmt.Println("connected to monitor")
	}

	return
}

func (teo *Teonet) proxyCall(address, command string, data []byte) (dataout []byte, err error) {
	// Send api request to teonet peer
	err = teo.ConnectTo(address)
	if err != nil {
		teo.Log().Debug.Println("can't connect teonet peer, err:", err)
		return
	}
	api, err := teo.NewAPIClient(address)
	if err != nil {
		teo.Log().Debug.Println("can't connect to api, err:", err)
		return
	}
	id, err := api.SendTo(command, data)
	if err != nil {
		teo.Log().Debug.Println("can't send api command, err:", err)
		return
	}
	teo.Log().Debug.Printf("send to %s cmd %s\n", address, command)
	dataout, err = api.WaitFrom(command, uint32(id))
	if err != nil {
		teo.Log().Debug.Println("can't get api data, err", err)
		return
	}
	teo.Log().Debug.Printf("got from %s cmd %s, data len: %d\n", address,
		command, len(dataout))

	return
}
