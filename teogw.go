// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teogw client connection and processing package
package teogw

import (
	"errors"
	"fmt"
	"time"

	"github.com/teonet-go/tru"
	"github.com/teonet-go/tru/teolog"
)

// Teogw structure and methods receiver
type Teogw struct {
	*tru.Tru
	log  *teolog.Teolog
	ch   *tru.Channel
	subs subscribe
}

// Create tru and start listen on port
func New(port int, log *teolog.Teolog, params ...interface{}) (t *Teogw, err error) {

	t = new(Teogw)
	t.subs.init()

	// Create server connection and start listen incominng packets
	t.Tru, err = tru.New(port, params...)
	if err != nil {
		return
	}
	if log == nil {
		log = teolog.New()
	}
	t.log = log

	return
}

// Create tru and start listen on port and connect to tru channel
func Connect(addr string, params ...interface{}) (t *Teogw, err error) {
	t, err = New(0, nil, params...)
	if err != nil {
		return
	}

	t.ch, err = t.Connect(addr, t.reader)

	return
}

// Send command to to teonet address
func (t *Teogw) Send(address, command string, data []byte) (pacid uint32, err error) {

	if t.ch == nil {
		err = errors.New("channel has not connected")
		return
	}

	var gwdata = TeogwData{0, address, command, data, ""}
	data, err = gwdata.MarshalBinary()
	if err != nil {
		return
	}

	id, err := t.ch.WriteTo(data)
	if err != nil {
		return
	}
	pacid = uint32(id)

	return
}

// Wait answer from teonet address by id or timeout
func (t *Teogw) Wait(address string, id uint32) (data []byte, err error) {

	select {
	case chandata := <-t.subscribe(address, id):
		data = chandata.data
		err = chandata.err
	case <-time.After(5 * time.Second):
		err = errors.New("can't get answer during timeout")
	}
	t.unsubscribe(address, id)

	return
}

// reader receive all tru channels packets
func (t *Teogw) reader(ch *tru.Channel, pac *tru.Packet, err error) (processed bool) {

	if err != nil {
		t.log.Debug.Println("got error in main reader:", err)
		return
	}

	var gwdata TeogwData

	err = gwdata.UnmarshalBinary(pac.Data())
	if err != nil {
		fmt.Println("wrong packet received:", err)
		return
	}

	c, exists := t.subs.get(gwdata.address, gwdata.id)
	if !exists {
		fmt.Println("wait subscription does not exists:", gwdata.address, gwdata.id)
		return
	}

	if len(gwdata.err) > 0 {
		err = errors.New(gwdata.err)
	} else {
		err = nil
	}
	c <- channelData{gwdata.data, err}

	return
}

// subscribe to answer
func (t *Teogw) subscribe(address string, id uint32) channel /* channel_read  */ {
	ch := make(channel)
	t.subs.set(address, id, ch)
	return ch
}

// unsubscribe from answer
func (t *Teogw) unsubscribe(address string, id uint32) (c channel /* channel_read */, exists bool) {
	c, exists = t.subs.get(address, id)
	if !exists {
		return
	}
	t.subs.del(address, id)

	return
}
