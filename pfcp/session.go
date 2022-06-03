// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"log"
	"net"
	"sort"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPSession struct {
	// isEstablished flag is used when PFCP Session Establishment Procedure has been completed
	// (can be initiated from the Local Entity or the Remote Peer, depending on kind of peer (UP/CP)
	isEstablished bool
	// association is used to send Request type PFCP Messages
	association api.PFCPAssociationInterface // XXX: use remoteFSEID to find the association from LocalEntity instead of storing an association
	// When Peer A send a message (M) to Peer B
	// M.PFCPHeader.SEID = B.LocalSEID() = A.RemoteSEID()
	// M.IPHeader.IP_DST = B.LocalIPAddress = A.RemoteIPAddress()
	localFseid  *ie.IE // local F-SEID
	remoteFseid *ie.IE // remote F-SEID, on UP function this is allocated at Setup() time
	// PDR Map allow to retrieve a specific PDR by its ID
	pdr pfcprule.PDRMap
	// sortedPDR is used to perform PDR finding using PDI Matching
	sortedPDR pfcprule.PDRs
	// FAR Map allow to retrieve a specific FAR by its ID
	far      pfcprule.FARMap
	atomicMu sync.Mutex // allows to perform atomic operations
}

// Create an EstablishedPFCPSession
// Use this function when a PFCP Session Establishment Request is received (UP case),
// or when the Entity want to send a PFCP Session Establishment Request (CP case).
func newEstablishedPFCPSession(association api.PFCPAssociationInterface, fseid, rseid *ie.IE, pdrs pfcprule.PDRMap, fars pfcprule.FARMap) (api.PFCPSessionInterface, error) {
	s := PFCPSession{
		isEstablished: false,
		association:   association,
		localFseid:    fseid, // local F-SEID
		remoteFseid:   rseid, // SEID present in FSEID ie send by remote peer
		pdr:           pdrs,
		far:           fars,
		sortedPDR:     make(pfcprule.PDRs, 0),
		atomicMu:      sync.Mutex{},
	}
	// sort PDRs
	for _, p := range pdrs {
		s.sortedPDR = append(s.sortedPDR, p)
	}
	sort.Sort(s.sortedPDR)
	if err := s.Setup(); err != nil {
		return nil, err
	}
	// Add to SessionFSEIDMap of LocalEntity
	s.association.LocalEntity().AddEstablishedPFCPSession(s)
	return s, nil
}

// Get local F-SEID of this session
// This value should be used when a session related message is received.
func (s PFCPSession) LocalFSEID() *ie.IE {
	return s.localFseid
}

// Get SEID part of local F-SEID
// This value should be used when a session related message is received.
func (s PFCPSession) LocalSEID() (api.SEID, error) {
	fseid, err := s.localFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

// Get IP Address part of local F-SEID
// This value should be used when a session related message is received.
func (s PFCPSession) LocalIPAddress() (net.IP, error) {
	fseid, err := s.localFseid.FSEID()
	if err != nil {
		return nil, err
	}
	switch {
	case fseid.HasIPv6():
		return fseid.IPv6Address, nil
	case fseid.HasIPv4():
		return fseid.IPv4Address, nil
	default:
		return nil, fmt.Errorf("Local IP Address not set")
	}
}

// Get remote F-SEID of this session
// This value should be used when a session related message is send.
func (s PFCPSession) RemoteFSEID() *ie.IE {
	return s.remoteFseid
}

// Get SEID part of remote F-SEID
// This value should be used when a session related message is send.
func (s PFCPSession) RemoteSEID() (api.SEID, error) {
	fseid, err := s.remoteFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

// Get IP Address part of remote F-SEID
// This value should be used when a session related message is send.
func (s PFCPSession) RemoteIPAddress() (net.IP, error) {
	fseid, err := s.remoteFseid.FSEID()
	if err != nil {
		return nil, err
	}
	switch {
	case fseid.HasIPv6():
		return fseid.IPv6Address, nil
	case fseid.HasIPv4():
		return fseid.IPv4Address, nil
	default:
		return nil, fmt.Errorf("Remote IP Address not set")
	}
}

// Returns PDRs sorted by Precedence
// For PDI checking, the checking order is:
// look first at the first item of the array,
// look last at the last item of the array.
func (s PFCPSession) GetPDRs() pfcprule.PDRs {
	s.atomicMu.Lock()
	defer s.atomicMu.Unlock()
	return s.sortedPDR
}

// Get FAR associated with this FARID
func (s PFCPSession) GetFAR(farid pfcprule.FARID) (*pfcprule.FAR, error) {
	if far, ok := s.far[farid]; ok {
		return far, nil
	}
	return nil, fmt.Errorf("No FAR with id", farid)
}

// Add PDRs to the session
// This is an internal function, not thread safe
func (s *PFCPSession) addPDRsUnsafe(pdrs pfcprule.PDRMap) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	for id, pdr := range pdrs {
		s.pdr[id] = pdr
		s.sortedPDR = append(s.sortedPDR, pdr)
	}
	sort.Sort(s.sortedPDR)
}

// Add FARs to the session
// This is an internal function, not thread safe
func (s *PFCPSession) addFARsUnsafe(fars pfcprule.FARMap) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	for id, far := range fars {
		s.far[id] = far
	}
}

// Add PDRs and FARs to the session
func (s PFCPSession) AddPDRsFARs(pdrs pfcprule.PDRMap, fars pfcprule.FARMap) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	s.atomicMu.Lock()
	defer s.atomicMu.Unlock()
	s.addPDRsUnsafe(pdrs)
	s.addFARsUnsafe(fars)
	// TODO: if isUserPlane() -> send the Session Modification Request
}

// Set the remote FSEID of a PFCPSession
// it must be used for next session related messages
//func (s PFCPSession) SetRemoteFSEID(FSEID *ie.IE) {
//	s.remoteFseid = FSEID
//XXX: change association to the right-one (unless XXX line 26 is fixed)
//     update sessionsmap in local entity
//}

// Setup function, either by:
// performing the PFCP Session Establishment Procedure (if CP function),
// or by doing nothing particular (if UP function) since
// the PFCP Session Establishment Procedure is already performed
func (s PFCPSession) Setup() error {
	if s.isEstablished {
		return fmt.Errorf("Session is already establihed")
	}
	switch {
	case s.association.LocalEntity().IsUserPlane():
		// Nothing more to do
		s.isEstablished = true
		return nil
	case s.association.LocalEntity().IsControlPlane():
		// Send PFCP Session Setup Request
		// first add to temporary map to avoid erroring after msg is send
		ies := make([]*ie.IE, 0)
		ies = append(ies, s.association.LocalEntity().NodeID())
		ies = append(ies, s.localFseid)
		for _, pdr := range pfcprule.NewCreatePDRs(s.pdr) {
			ies = append(ies, pdr)
		}
		for _, far := range pfcprule.NewCreateFARs(s.far) {
			ies = append(ies, far)
		}

		msg := message.NewSessionEstablishmentRequest(0, 0, 0, 0, 0, ies...)
		resp, err := s.association.Send(msg)
		if err != nil {
			return err
		}
		ser, ok := resp.(*message.SessionEstablishmentResponse)
		if !ok {
			log.Printf("got unexpected message: %s\n", resp.MessageTypeName())
		}
		s.remoteFseid = ser.UPFSEID
		s.isEstablished = true
		return nil
	default:
		return fmt.Errorf("Local PFCP entity is not a CP or a UP function")
	}
	return nil
}
