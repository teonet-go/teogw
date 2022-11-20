// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tru connection and processing module

package main

import (
	"github.com/teonet-go/teogw"
	"github.com/teonet-go/teonet"
	"github.com/teonet-go/tru"
	"github.com/teonet-go/tru/teolog"
)

type Tru struct {
	*teogw.Teogw
	teo *teonet.Teonet
	log *teolog.Teolog
}

// Connect and start Tru
func newTru(teo *teonet.Teonet) (t *Tru, err error) {

	t = new(Tru)
	t.teo = teo
	t.log = teo.Log()

	// // Create server connection and start listen incominng packets
	// t.Tru, err = tru.New(params.tru_port, t.reader, tru.Stat(params.tru_stat),
	// 	/* tru.Hotkey(*hotkey), */ log, params.loglevel,
	// 	teolog.Logfilter(params.logfilter))
	// if err != nil {
	// 	return
	// }
	//

	t.Teogw, err = teogw.New(params.tru_port, t.log, t.reader, tru.Stat(params.tru_stat),
		/* tru.Hotkey(*hotkey), */ t.log, params.loglevel,
		teolog.Logfilter(params.logfilter))

	return
}

// Reader read packets from connected peers
func (t *Tru) reader(ch *tru.Channel, pac *tru.Packet, err error) (processed bool) {

	if err != nil {
		// t.log.Debug.Println("got error in main reader:", err)
		return
	}
	// t.log.Debugv.Printf("got %d byte from %s, id %d, data len: %d\n",
	// 	pac.Len(), ch.Addr().String(), pac.ID(), len(pac.Data()))

	var gw teogw.TeogwData
	err = gw.UnmarshalBinary(pac.Data())
	if err != nil {
		t.log.Debug.Println("got wrong packet")
		return
	}
	t.log.Debugv.Println("got teogw request", gw.Address(), gw.Command())
	gw.SetID(uint32(pac.ID()))

	go func(ch *tru.Channel, gw teogw.TeogwData) {

		var err error

		// Send answer before return
		defer func() {
			if err != nil {
				gw.SetError(err)
			}
			t.sendAnswer(ch, &gw)
		}()

		// Send api request to teonet peer
		err = t.teo.ConnectTo(gw.Address())
		if err != nil {
			t.log.Debug.Println("can't connect teonet peer, err:", err)
			return
		}
		api, err := t.teo.NewAPIClient(gw.Address())
		if err != nil {
			t.log.Debug.Println("can't connect to api, err:", err)
			return
		}
		id, err := api.SendTo(gw.Command(), gw.Data())
		if err != nil {
			t.log.Debug.Println("can't send api command, err:", err)
			return
		}
		data, err := api.WaitFrom(gw.Command(), uint32(id))
		if err != nil {
			t.log.Debug.Println("can't get api data, err", err)
			return
		}
		gw.SetData(data)

	}(ch, gw)

	return
}

func (t *Tru) sendAnswer(ch *tru.Channel, gw *teogw.TeogwData) (err error) {
	data, err := gw.MarshalBinary()
	if err != nil {
		return
	}
	_, err = ch.WriteTo(data)
	return
}
