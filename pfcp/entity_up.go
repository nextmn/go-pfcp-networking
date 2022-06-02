// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"log"

	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntityUP struct {
	PFCPEntity
}

func NewPFCPSEntityUP(nodeID string) *PFCPEntityUP {
	e := PFCPEntityUP{PFCPEntity: NewPFCPEntity(nodeID, "UP")}
	err := e.initDefaultHandlers()
	if err != nil {
		log.Println(err)
	}
	return &e
}

func (e *PFCPEntityUP) initDefaultHandlers() error {
	if err := e.AddHandler(message.MsgTypeAssociationSetupRequest, handleAssociationSetupRequest); err != nil {
		return err
	}
	if err := e.AddHandler(message.MsgTypeSessionEstablishmentRequest, handleSessionEstablishmentRequest); err != nil {
		return err
	}
	return nil
}

//func (e *PFCPEntityUP) GetLocalSessions() PFCPSessionMapSEID {
//	// TODO: Store Session global map directly in the entity and only store array of SEIDs in association
//	var s PFCPSessionMapSEID
//	for _, a := range e.associations {
//		for k, v := range a.GetSessions() {
//			s[k] = v
//		}
//	}
//	return s
//}

//func (e *PFCPEntityUP) GetPFCPSessions() []*PFCPSession {
//	sessions := make([]*PFCPSession, 0)
//	for _, a := range e.associations {
//		as := a.GetSessions()
//		for _, s := range as {
//			sessions = append(sessions, s)
//		}
//	}
//	return sessions
//}
