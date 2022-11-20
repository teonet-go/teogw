// Copyright 2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Subscribe structure and methods

package teogw

import (
	"fmt"
	"sync"
)

// Subscribe to answer type
type subscribe struct {
	m   map[string]channel
	mut *sync.RWMutex
}

// Subscribe map (value) channel type
type channel chan channelData
type channelData struct {
	data []byte
	err  error
}

// Initialize subscribe object
func (s *subscribe) init() {
	s.m = make(map[string]channel)
	s.mut = new(sync.RWMutex)
}

// Create key
func (s *subscribe) key(address string, id uint32) string {
	return fmt.Sprintf("%s,%d", address, id)
}

// Get subscribed channel
func (s *subscribe) get(address string, id uint32) (c channel, exists bool) {
	s.mut.RLock()
	defer s.mut.RUnlock()

	key := s.key(address, id)
	c, exists = s.m[key]
	return
}

// Set subscribe channel
func (s *subscribe) set(address string, id uint32, c channel) {
	s.mut.Lock()
	defer s.mut.Unlock()

	key := s.key(address, id)
	s.m[key] = c
}

// Delete subscribe
func (s *subscribe) del(address string, id uint32) {
	s.mut.Lock()
	defer s.mut.Unlock()

	key := s.key(address, id)
	delete(s.m, key)
}
