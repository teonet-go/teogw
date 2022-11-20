// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teonet connection and processing module

package main

import (
	"flag"
	"fmt"

	"github.com/teonet-go/teomon"
	"github.com/teonet-go/teonet"
)

type Parameters struct {
	appShortName string
	port         int
	tru_port     int
	stat         bool
	tru_stat     bool
	hotkey       bool
	loglevel     string
	logfilter    string
}

var params Parameters

// Connect and start Teonet
func Teonet() (teo *teonet.Teonet, err error) {

	// Parse application command line parameters
	flag.StringVar(&params.appShortName, "name", appShort, "application short name")
	flag.IntVar(&params.port, "p", 0, "local port")
	flag.IntVar(&params.tru_port, "tp", 0, "tru local port")
	flag.BoolVar(&params.stat, "stat", false, "show trudp statistic")
	flag.BoolVar(&params.tru_stat, "tru-stat", false, "show trudp statistic")
	flag.BoolVar(&params.hotkey, "hotkey", false, "run hotkey meny")
	flag.StringVar(&params.loglevel, "loglevel", "NONE", "log level")
	flag.StringVar(&params.logfilter, "logfilter", "", "log filter")
	flag.StringVar(&monitor, "monitor", "", "monitor address")
	flag.Parse()

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
