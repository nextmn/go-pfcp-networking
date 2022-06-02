// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

type PFCPEntityCP struct {
	PFCPEntity
}

func NewPFCPEntityCP(nodeID string) *PFCPEntityCP {
	e := PFCPEntityCP{PFCPEntity: NewPFCPEntity(nodeID, "CP")}
	return &e
}

//func (e *PFCPEntityCP) GetLocalSessions() PFCPSessionMapSEID {
//	// TODO: Store Session global map directly in the entity and only store array of SEIDs in association
//	var s PFCPSessionMapSEID
//	for _, a := range e.associations {
//		for k, v := range a.GetSessions() {
//			s[k] = v
//		}
//	}
//	return s
//}
