// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Teogw structure and methods

package teogw

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/kirill-scherba/bslice"
	"github.com/teonet-go/teowebrtc_server"
)

// Teogw packet data
type TeogwData struct {
	ID      uint32 `json:"id"`
	Address string `json:"address"`
	Command string `json:"command"`
	Data    []byte `json:"data"`
	Err     string `json:"err"`
}

// Set id
func (gw *TeogwData) SetID(id uint32) {
	gw.ID = id
}

// Set data
func (gw *TeogwData) SetCommand(command string) {
	gw.Command = command
}

// Set data
func (gw *TeogwData) SetData(data []byte) {
	gw.Data = data
}

// Set error
func (gw *TeogwData) SetError(err error) {
	gw.Err = err.Error()
}

// Get ID
func (gw TeogwData) GetID() uint32 {
	return gw.ID
}

// Get address
func (gw TeogwData) GetAddress() string {
	return gw.Address
}

// Get command
func (gw TeogwData) GetCommand() string {
	return gw.Command
}

// Get data
func (gw TeogwData) GetData() []byte {
	return gw.Data
}

// MarshalBinary marshal Teogw binary packet
func (gw *TeogwData) MarshalBinary() (out []byte, err error) {

	var b bslice.ByteSlice
	buf := new(bytes.Buffer)
	le := binary.LittleEndian

	// b.WriteSlice(buf, []byte(gw.id))
	binary.Write(buf, le, gw.ID)
	b.WriteSlice(buf, []byte(gw.Address))
	b.WriteSlice(buf, []byte(gw.Command))
	b.WriteSlice(buf, []byte(gw.Data))
	b.WriteSlice(buf, []byte(gw.Err))

	out = buf.Bytes()
	return
}

// UnmarshalBinary unmarshal Teogw binary packet
func (gw *TeogwData) UnmarshalBinary(data []byte) (err error) {

	var b bslice.ByteSlice
	buf := bytes.NewBuffer(data)
	le := binary.LittleEndian

	// gw.id, err = b.ReadString(buf)
	binary.Read(buf, le, &gw.ID)
	if err != nil {
		return
	}

	gw.Address, err = b.ReadString(buf)
	if err != nil {
		return
	}

	gw.Command, err = b.ReadString(buf)
	if err != nil {
		return
	}

	gw.Data, err = b.ReadSlice(buf)
	if err != nil {
		return
	}

	gw.Err, err = b.ReadString(buf)

	return
}

// Unmarshal Json and return TeogwData
func (gw TeogwData) UnmarshalJson(data []byte) (teowebrtc_server.TeogwData, error) {
	var request = TeogwData{}
	err := json.Unmarshal(data, &request)
	return request, err
}

// Marshal TeogwData to Json
func (gw TeogwData) MarshalJson(
	gwData teowebrtc_server.TeogwData,
	command string, inData []byte, inErr error) ([]byte, error) {

	var request = TeogwData{
		ID:      gwData.GetID(),
		Address: gwData.GetAddress(),
		Command: command,
		Data:    inData,
		Err: func() (errStr string) {
			if inErr != nil {
				errStr = inErr.Error()
			}
			return
		}(),
	}

	return json.Marshal(request)
}
