// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"context"
	"net/netip"

	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPEntityInterface interface {
	ListenAddr() netip.Addr
	IsUserPlane() bool
	IsControlPlane() bool
	NodeID() *ie.IE
	RecoveryTimeStamp() *ie.IE
	NewEstablishedPFCPAssociation(nodeID *ie.IE) (association PFCPAssociationInterface, err error)
	RemovePFCPAssociation(association PFCPAssociationInterface) error
	GetPFCPAssociation(nid string) (association PFCPAssociationInterface, err error)
	GetPFCPSessions() []PFCPSessionInterface
	GetPFCPSession(localIP string, seid SEID) (PFCPSessionInterface, error)
	AddEstablishedPFCPSession(session PFCPSessionInterface) error
	LogPFCPRules()
	Options() EntityOptionsInterface
	WaitReady(ctx context.Context) error
}
