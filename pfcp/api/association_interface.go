// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"github.com/wmnsk/go-pfcp/ie"
)

type SEID = uint64
type PFCPAssociationInterface interface {
	PFCPPeerInterface
	SetupInitiatedByCP() error
	GetNextSEID() SEID
	CreateSession(remoteFseid *ie.IE, pdrs PDRMapInterface, fars FARMapInterface) (session PFCPSessionInterface, err error)
}
