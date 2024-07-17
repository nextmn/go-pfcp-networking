// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"log"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntityUP struct {
	PFCPEntity
}

func NewPFCPEntityUP(nodeID string, listenAddr string) *PFCPEntityUP {
	return NewPFCPEntityUPWithOptions(nodeID, listenAddr, EntityOptions{})
}

func NewPFCPEntityUPWithOptions(nodeID string, listenAddr string, options api.EntityOptionsInterface) *PFCPEntityUP {
	e := PFCPEntityUP{PFCPEntity: NewPFCPEntity(nodeID, listenAddr, "UP", options)}
	err := e.initDefaultHandlers()
	if err != nil {
		log.Println(err)
	}
	return &e
}

func (e *PFCPEntityUP) initDefaultHandlers() error {
	if err := e.AddHandler(message.MsgTypeAssociationSetupRequest, DefaultAssociationSetupRequestHandler); err != nil {
		return err
	}
	if err := e.AddHandler(message.MsgTypeSessionEstablishmentRequest, DefaultSessionEstablishmentRequestHandler); err != nil {
		return err
	}
	if err := e.AddHandler(message.MsgTypeSessionModificationRequest, DefaultSessionModificationRequestHandler); err != nil {
		return err
	}
	return nil
}
