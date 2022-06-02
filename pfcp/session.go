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
	isEstablished bool
	association   api.PFCPAssociationInterface // is used to send Request type PFCP Messages
	// When Peer A send a message (M) to Peer B
	// M.PFCPHeader.SEID = B.LocalSEID() = A.RemoteSEID()
	// M.IPHeader.IP_DST = B.LocalIPAddress = A.RemoteIPAddress()
	localFseid  *ie.IE // local F-SEID
	remoteFseid *ie.IE // remote F-SEID
	pdr         map[uint16]*pfcprule.PDR
	far         map[uint32]*pfcprule.FAR
	sortedPDR   pfcprule.PDRs
	atomicMu    sync.Mutex // allows to perform atomic operations
}

func newEstablishedPFCPSession(association api.PFCPAssociationInterface, fseid, rseid *ie.IE, pdrs map[uint16]*pfcprule.PDR, fars map[uint32]*pfcprule.FAR) (api.PFCPSessionInterface, error) {
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
	return s, nil
}

func (s PFCPSession) LocalFSEID() *ie.IE {
	return s.localFseid
}

func (s PFCPSession) LocalSEID() (uint64, error) {
	fseid, err := s.localFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

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

func (s PFCPSession) RemoteFSEID() *ie.IE {
	return s.remoteFseid
}

func (s PFCPSession) RemoteSEID() (uint64, error) {
	fseid, err := s.remoteFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

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

func (s PFCPSession) GetPDRs() pfcprule.PDRs {
	s.atomicMu.Lock()
	defer s.atomicMu.Unlock()
	return s.sortedPDR
}

func (s PFCPSession) GetFAR(farid uint32) (*pfcprule.FAR, error) {
	if far, ok := s.far[farid]; ok {
		return far, nil
	}
	return nil, fmt.Errorf("No far with id", farid)
}

func (s *PFCPSession) addPDRsUnsafe(pdrs map[uint16]*pfcprule.PDR) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	for id, pdr := range pdrs {
		s.pdr[id] = pdr
		s.sortedPDR = append(s.sortedPDR, pdr)
	}
	sort.Sort(s.sortedPDR)

}
func (s *PFCPSession) addFARsUnsafe(fars map[uint32]*pfcprule.FAR) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	for id, far := range fars {
		s.far[id] = far
	}
}

func (s PFCPSession) AddPDRsFARs(pdrs map[uint16]*pfcprule.PDR, fars map[uint32]*pfcprule.FAR) {
	// Transactions must be atomic to avoid having a PDR referring to a deleted FAR / not yet created FAR
	s.atomicMu.Lock()
	defer s.atomicMu.Unlock()
	s.addPDRsUnsafe(pdrs)
	s.addFARsUnsafe(fars)
}

// Set the remote FSEID of a PFCPSession
func (s PFCPSession) SetRemoteFSEID(FSEID *ie.IE) {

}

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
		tmpPDR := make(map[uint16]*pfcprule.PDR)
		for _, pdr := range s.pdr {
			id, err := pdr.ID()
			if err != nil {
				return err
			}
			tmpPDR[id] = pdr
		}
		tmpFAR := make(map[uint32]*pfcprule.FAR)
		for _, far := range s.far {
			id, err := far.ID()
			if err != nil {
				return err
			}
			tmpFAR[id] = far
		}
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
		s.AddPDRsFARs(tmpPDR, tmpFAR)
		s.isEstablished = true
		return nil
	default:
		return fmt.Errorf("Local PFCP entity is not a CP or a UP function")
	}
	return nil
}
