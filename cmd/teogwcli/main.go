// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teogw client sample apllication
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/teonet-go/teogw"
	"github.com/teonet-go/teonet"
)

func main() {

	var n int
	var peers string
	flag.IntVar(&n, "n", 1, "number of requests")
	flag.StringVar(&peers, "peers", "", "comma delimited teoapi peers")
	flag.Parse()
	teonet.CheckRequeredParams("peers")

	addrs := strings.Split(peers, ",")

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Println("start")
	start := time.Now()

	// Connect to Teogw Server
	t, err := teogw.Connect("localhost:7701")
	if err != nil {
		log.Fatalln("can't connect to teogw, error:", err)
	}
	defer t.Close()
	log.Println("connected to teogw")

	for i := 0; i < n; i++ {

		// Send echo to teoapi
		const cmd = "hello"
		addr := addrs[i%2]
		id, err := t.Send(addr, cmd, []byte(fmt.Sprintf("Mememe-%d!", i)))
		if err != nil {
			log.Println("can't send command, error:", err)
			continue
		}
		log.Printf("send id: %d, to: %s", id, addr)

		// Wait answer from teoapi
		data, err := t.Wait(addr, id)
		if err != nil {
			log.Println("can't got answer to command, error:", err)
			continue
		}
		log.Printf("got answer from: %s, id: %d, dsta: %s\n", addr, id, data)
	}

	log.Println("it took time:", time.Since(start))
}
