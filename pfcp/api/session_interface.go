// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"net"

	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPSessionInterface interface {
	LocalFSEID() *ie.IE
	LocalSEID() (SEID, error)
	LocalIPAddress() (net.IP, error)
	RemoteFSEID() *ie.IE
	RemoteSEID() (SEID, error)
	RemoteIPAddress() (net.IP, error)
	GetPDRs() pfcprule.PDRs
	GetFAR(farid uint32) (*pfcprule.FAR, error)
	AddPDRsFARs(pdrs pfcprule.PDRMap, fars pfcprule.FARMap)
	//	SetRemoteFSEID(FSEID *ie.IE)
	Setup() error
}
