// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"sync"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
)

// XXX Delete old sessions instead of just creating new ones
type sessionsMapSEID = map[api.SEID]api.PFCPSessionInterface
type sessionsMapFSEID = map[string]sessionsMapSEID
type SessionsMap struct {
	sessions   sessionsMapFSEID
	muSessions sync.RWMutex
}

// Add a session to the map
func (sm *SessionsMap) Add(session api.PFCPSessionInterface) error {
	sm.muSessions.Lock()
	defer sm.muSessions.Unlock()
	// Get splitted F-SEID
	localIPAddr, err := session.LocalIPAddress() // XXX: handle case where both ip6 and ip4 are set
	if err != nil {
		return err
	}
	localIP := localIPAddr.String()
	localSEID, err := session.LocalSEID()
	if err != nil {
		return err
	}
	// Create submap if first session with this localIP
	if _, exists := sm.sessions[localIP]; !exists {
		sm.sessions[localIP] = make(sessionsMapSEID, 0)
	}
	// Add session
	sm.sessions[localIP][localSEID] = session
	return nil
}

// Create a new SessionMap
func NewSessionsMap() *SessionsMap {
	return &SessionsMap{
		sessions:   make(sessionsMapFSEID, 0),
		muSessions: sync.RWMutex{},
	}
}

// Returns pfcpsessions in an array
func (sm *SessionsMap) GetPFCPSessions() []api.PFCPSessionInterface {
	sm.muSessions.RLock()
	defer sm.muSessions.RUnlock()
	sessions := make([]api.PFCPSessionInterface, 0)
	for _, byseid := range sm.sessions {
		for _, session := range byseid {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// Returns a PFCP Session by its FSEID
func (sm *SessionsMap) GetPFCPSession(localIP string, seid api.SEID) (api.PFCPSessionInterface, error) {
	sm.muSessions.RLock()
	defer sm.muSessions.RUnlock()
	if sessions, ipexists := sm.sessions[localIP]; ipexists {
		if session, sessionexists := sessions[seid]; sessionexists {
			return session, nil
		} else {
			return nil, fmt.Errorf("session not found: wrong SEID")
		}
	} else {
		return nil, fmt.Errorf("session not found: wrong IP")
	}
}
