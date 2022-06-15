// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"net"

	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPSessionInterface interface {
	LocalFSEID() *ie.IE
	LocalSEID() (SEID, error)
	LocalIPAddress() (net.IP, error)
	RemoteFSEID() *ie.IE
	RemoteSEID() (SEID, error)
	RemoteIPAddress() (net.IP, error)
	GetSortedPDRIDs() []PDRID
	GetPDR(pdrid PDRID) (PDRInterface, error)
	GetFAR(farid FARID) (FARInterface, error)
	AddUpdatePDRsFARs(createpdrs PDRMapInterface, createfars FARMapInterface, updatepdr PDRMapInterface, updatefars FARMapInterface) error
	//	SetRemoteFSEID(FSEID *ie.IE)
	Setup() error
	ForeachUnsortedPDR(f func(pdr PDRInterface) error) error

	// Must be called before getting PDRIDs, PDR, and FARs in one operation
	// to ensure FARs are up-to-date with PDRs
	RLock()
	RUnlock()
}
