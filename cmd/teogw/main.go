// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teogw teonet application get request by tru, execute it in teonet and send
// responce back by tru. The tequest format is: {id, address, command, data}
// where:
//
//	id - client id
//	address - teonet application address
//	command - teonet application api command
//	data - teonet application api command data
package main

import (
	"flag"
	"log"
	"time"

	"github.com/teonet-go/teonet"
)

const (
	appShort   = "teogw"
	appName    = "Teonet gateway application"
	appLong    = ""
	appVersion = "0.6.4"
)

var appStartTime = time.Now()
var monitor string

type Parameters struct {
	appShortName string
	port         int
	stat         bool
	tru_stat     bool
	hotkey       bool
	loglevel     string
	logfilter    string

	tru_port int

	signalAddr    string
	signalAddrTls string
}

var params Parameters

func main() {

	// Application logo
	teonet.Logo(appName, appVersion)

	// Parse application command line parameters
	flag.StringVar(&params.appShortName, "name", appShort, "application short name")
	flag.IntVar(&params.port, "p", 0, "local port")
	flag.BoolVar(&params.stat, "stat", false, "show trudp statistic")
	flag.BoolVar(&params.tru_stat, "tru-stat", false, "show trudp statistic")
	flag.BoolVar(&params.hotkey, "hotkey", false, "run hotkey meny")
	flag.StringVar(&params.loglevel, "loglevel", "NONE", "log level")
	flag.StringVar(&params.logfilter, "logfilter", "", "log filter")
	flag.StringVar(&monitor, "monitor", "", "monitor address")

	flag.IntVar(&params.tru_port, "tp", 7701, "tru local port")

	flag.StringVar(&params.signalAddr, "signal-addr", ":8081", "webrtc http signal service address")
	flag.StringVar(&params.signalAddrTls, "signal-addr-ssl", ":8082", "webrtc https signal service address")

	flag.Parse()

	// Connect and start Teonet
	teo, err := newTeonet()
	if err != nil {
		log.Fatalln("can't strat teonet, error:", err)
	}
	defer teo.Close()

	// Connect and start Tru proxy
	t, err := newTru(teo)
	if err != nil {
		log.Fatalln("can't strat tru, error:", err)
	}
	defer t.Close()

	// Connect and start WebRTC proxy
	newWebRTC(teo)

	select {}
}
