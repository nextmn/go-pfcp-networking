// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"log/slog"
	"sync"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
)

// SessionIDPool is a generator of session IDs
type SessionIDPool struct {
	currentSessionID api.SEID
	muSessionID      sync.Mutex
}

// Create a SessionIDPool
func NewSessionIDPool() *SessionIDPool {
	return &SessionIDPool{
		currentSessionID: 1,
		muSessionID:      sync.Mutex{},
	}
}

// Get next id available in SessionIDPool
func (pool *SessionIDPool) GetNext() api.SEID {
	//FIXME: add a Context
	// XXX: make session ID non-predictable?
	pool.muSessionID.Lock()
	defer pool.muSessionID.Unlock()
	id := pool.currentSessionID
	pool.currentSessionID = id + 1
	slog.Debug("Returning next Session ID", "next-session-id", id)
	return id
}
