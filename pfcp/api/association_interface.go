// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
)

type SEID = uint64
type PFCPAssociationInterface interface {
	PFCPPeerInterface
	SetupInitiatedByCP() error
	GetNextSEID() SEID
	CreateSession(remoteFseid *ie.IE, pdrs pfcprule.PDRs, fars pfcprule.FARs) (session PFCPSessionInterface, err error)
}
