// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"log"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
)

// SessionIDPool is a generator of session IDs
type SessionIDPool struct {
	currentSessionID api.SEID
	muSessionID      sync.Mutex
}

// Create a SessionIDPool
func NewSessionIDPool() SessionIDPool {
	return SessionIDPool{
		currentSessionID: 0,
		muSessionID:      sync.Mutex{},
	}
}

// Get next id available in SessionIDPool
func (pool *SessionIDPool) GetNext() api.SEID {
	pool.muSessionID.Lock()
	defer pool.muSessionID.Unlock()
	id := pool.currentSessionID
	pool.currentSessionID = id + 1
	log.Println("Returning next Session ID:", id)
	return id
}
