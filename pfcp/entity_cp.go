// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

type PFCPEntityCP struct {
	PFCPEntity
}

func NewPFCPEntityCP(nodeID string, listenAddr string) *PFCPEntityCP {
	return NewPFCPEntityCPWithOptions(nodeID, listenAddr, EntityOptions{})
}

func NewPFCPEntityCPWithOptions(nodeID string, listenAddr string, options EntityOptions) *PFCPEntityCP {
	return &PFCPEntityCP{PFCPEntity: NewPFCPEntity(nodeID, listenAddr, "CP", options)}
}
