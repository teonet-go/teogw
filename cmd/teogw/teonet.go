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

// Connect and start Teonet
func Teonet() (teo *teonet.Teonet, err error) {

	// Create Teonet connector
	teo, err = teonet.New(
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
