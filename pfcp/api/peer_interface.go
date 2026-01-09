// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPPeerInterface interface {
	IsRunning() bool
	Close() error
	Send(msg message.Message) (m message.Message, err error)
	IsAlive() (res bool, err error)
	NodeID() *ie.IE
	IsUserPlane() bool
	IsControlPlane() bool
	LocalEntity() PFCPEntityInterface
	NewEstablishedPFCPAssociation() (PFCPAssociationInterface, error)
}
