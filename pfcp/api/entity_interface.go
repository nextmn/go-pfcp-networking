// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"net"

	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPEntityInterface interface {
	IsUserPlane() bool
	IsControlPlane() bool
	NodeID() *ie.IE
	RecoveryTimeStamp() *ie.IE
	NewEstablishedPFCPAssociation(nodeID *ie.IE) (association PFCPAssociationInterface, err error)
	RemovePFCPAssociation(association PFCPAssociationInterface) error
	GetPFCPAssociation(nid string) (association PFCPAssociationInterface, err error)
	SendTo(msg []byte, dst net.Addr) error
	GetPFCPSessions() []PFCPSessionInterface
	GetPFCPSession(localIP string, seid SEID) (PFCPSessionInterface, error)
	UpdatePFCPAssociation(association PFCPAssociationInterface) error
	AddEstablishedPFCPSession(session PFCPSessionInterface) error
}
