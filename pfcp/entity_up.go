// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"net/netip"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/nextmn/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntityUP struct {
	PFCPEntity
}

func NewPFCPEntityUP(nodeID string, listenAddr netip.Addr) *PFCPEntityUP {
	return NewPFCPEntityUPWithOptions(nodeID, listenAddr, EntityOptions{})
}

func NewPFCPEntityUPWithOptions(nodeID string, listenAddr netip.Addr, options api.EntityOptionsInterface) *PFCPEntityUP {
	return &PFCPEntityUP{PFCPEntity: NewPFCPEntity(nodeID, listenAddr, "UP", newDefaultPFCPEntityUPHandlers(), options)}
}

func newDefaultPFCPEntityUPHandlers() map[pfcputil.MessageType]PFCPMessageHandler {
	m := newDefaultPFCPEntityHandlers()
	m[message.MsgTypeAssociationSetupRequest] = DefaultAssociationSetupRequestHandler
	m[message.MsgTypeSessionEstablishmentRequest] = DefaultSessionEstablishmentRequestHandler
	m[message.MsgTypeSessionModificationRequest] = DefaultSessionModificationRequestHandler
	return m
}
