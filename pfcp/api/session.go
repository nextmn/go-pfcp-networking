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
	LocalSEID() (uint64, error)
	LocalIPAddress() (net.IP, error)
	RemoteFSEID() *ie.IE
	RemoteSEID() (uint64, error)
	RemoteIPAddress() (net.IP, error)
	GetPDRs() pfcprule.PDRs
	GetFAR(farid uint32) (*pfcprule.FAR, error)
	AddPDRsFARs(pdrs map[uint16]*pfcprule.PDR, fars map[uint32]*pfcprule.FAR)
	SetRemoteFSEID(FSEID *ie.IE)
	Setup() error
}
