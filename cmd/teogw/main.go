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
	"log"
	"time"

	"github.com/teonet-go/teonet"
)

const (
	appShort   = "teogw"
	appName    = "Teonet gateway application"
	appLong    = ""
	appVersion = "0.6.0"
)

var appStartTime = time.Now()
var monitor string

func main() {

	// Application logo
	teonet.Logo(appName, appVersion)

	// Connect and start Teonet
	teo, err := Teonet()
	if err != nil {
		log.Fatalln("can't strat teonet, error:", err)
	}
	defer teo.Close()

	// Connect and start Tru
	t, err := newTru(teo)
	if err != nil {
		log.Fatalln("can't strat tru, error:", err)
	}
	defer t.Close()

	select {}
}
