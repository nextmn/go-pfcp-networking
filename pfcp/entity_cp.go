// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"net/netip"
)

type PFCPEntityCP struct {
	PFCPEntity
}

func NewPFCPEntityCP(nodeID string, listenAddr netip.Addr, options ...entityOption) *PFCPEntityCP {
	return &PFCPEntityCP{NewPFCPEntity(nodeID, listenAddr, "CP", nil, options...)}
}
