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

func NewPFCPEntityUP(nodeID string) *PFCPEntityUP {
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
	if err := e.AddHandler(message.MsgTypeSessionModificationRequest, handleSessionModificationRequest); err != nil {
		return err
	}
	return nil
}
