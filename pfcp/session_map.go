// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
)

type sessionsMapSEID = map[uint64]api.PFCPSessionInterface
type sessionsMapFSEID = map[string]*sessionsMapSEID
type SessionsMap struct {
	sessions   sessionsMapFSEID
	muSessions sync.Mutex
}

func (sm *SessionsMap) Add(session api.PFCPSessionInterface) {
	//TODO
}
