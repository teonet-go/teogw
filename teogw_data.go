// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teogw structure and methods

package teogw

import (
	"bytes"
	"encoding/binary"

	"github.com/kirill-scherba/bslice"
)

// Teogw packet data
type TeogwData struct {
	id      uint32
	address string
	command string
	data    []byte
	err     string
}

// Set id
func (gw *TeogwData) SetID(id uint32) {
	gw.id = id
}

// Set data
func (gw *TeogwData) SetData(data []byte) {
	gw.data = data
}

// Set error
func (gw *TeogwData) SetError(err error) {
	gw.err = err.Error()
}

// Get address
func (gw *TeogwData) Address() string {
	return gw.address
}

// Get command
func (gw *TeogwData) Command() string {
	return gw.command
}

// Get data
func (gw *TeogwData) Data() []byte {
	return gw.data
}

// MarshalBinary marshal Teogw binary packet
func (gw *TeogwData) MarshalBinary() (out []byte, err error) {

	var b bslice.ByteSlice
	buf := new(bytes.Buffer)
	le := binary.LittleEndian

	// b.WriteSlice(buf, []byte(gw.id))
	binary.Write(buf, le, gw.id)
	b.WriteSlice(buf, []byte(gw.address))
	b.WriteSlice(buf, []byte(gw.command))
	b.WriteSlice(buf, []byte(gw.data))
	b.WriteSlice(buf, []byte(gw.err))

	out = buf.Bytes()
	return
}

// UnmarshalBinary unmarshal Teogw binary packet
func (gw *TeogwData) UnmarshalBinary(data []byte) (err error) {

	var b bslice.ByteSlice
	buf := bytes.NewBuffer(data)
	le := binary.LittleEndian

	// gw.id, err = b.ReadString(buf)
	binary.Read(buf, le, &gw.id)
	if err != nil {
		return
	}

	gw.address, err = b.ReadString(buf)
	if err != nil {
		return
	}

	gw.command, err = b.ReadString(buf)
	if err != nil {
		return
	}

	gw.data, err = b.ReadSlice(buf)
	if err != nil {
		return
	}

	gw.err, err = b.ReadString(buf)

	return
}
