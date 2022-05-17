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

	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPSessionMapSEID = map[uint64]*PFCPSession
type RemotePFCPSessionMapSEID = map[uint64]*RemotePFCPSession

type PFCPSession struct {
	localFseid  *ie.IE // local F-SEID
	remoteFseid *ie.IE // remote F-SEID
	pdr         map[uint16]*pfcprule.PDR
	far         map[uint32]*pfcprule.FAR
	sortedPDR   pfcprule.PDRs
	mu          sync.Mutex
}

func NewPFCPSession(fseid, rseid *ie.IE) PFCPSession {
	return PFCPSession{
		localFseid:  fseid, // local F-SEID
		remoteFseid: rseid, // SEID present in FSEID ie send by remote peer
		pdr:         make(map[uint16]*pfcprule.PDR),
		far:         make(map[uint32]*pfcprule.FAR),
		sortedPDR:   make(pfcprule.PDRs, 0),
		mu:          sync.Mutex{},
	}
}

func (s *PFCPSession) LocalFSEID() *ie.IE {
	return s.localFseid
}

func (s *PFCPSession) LocalSEID() (uint64, error) {
	fseid, err := s.localFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

func (s *PFCPSession) LocalIPAddress() (net.IP, error) {
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

func (s *PFCPSession) RemoteFSEID() *ie.IE {
	return s.remoteFseid
}

func (s *PFCPSession) RemoteSEID() (uint64, error) {
	fseid, err := s.remoteFseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

func (s *PFCPSession) RemoteIPAddress() (net.IP, error) {
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

func (s *PFCPSession) GetPDRs() pfcprule.PDRs {
	return s.sortedPDR
}

func (s *PFCPSession) GetFAR(farid uint32) (*pfcprule.FAR, error) {
	if far, ok := s.far[farid]; ok {
		return far, nil
	}
	return nil, fmt.Errorf("No far with id", farid)
}

func (s *PFCPSession) AddPDRs(pdrs map[uint16]*pfcprule.PDR) {
	s.mu.Lock()
	for id, pdr := range pdrs {
		s.pdr[id] = pdr
		s.sortedPDR = append(s.sortedPDR, pdr)
	}
	sort.Sort(s.sortedPDR)
	s.mu.Unlock()

}
func (s *PFCPSession) AddFARs(fars map[uint32]*pfcprule.FAR) {
	s.mu.Lock()
	for id, far := range fars {
		s.far[id] = far
	}
	s.mu.Unlock()
}

type RemotePFCPSession struct {
	PFCPSession
	association *PFCPAssociation
}

func NewRemotePFCPSession(localFseid *ie.IE, association *PFCPAssociation) RemotePFCPSession {
	return RemotePFCPSession{
		PFCPSession: NewPFCPSession(localFseid, nil), // Remote SEID is initialized when a Session Establishment Response is received
		association: association,
	}
}

func (s *RemotePFCPSession) Start(pdrs []*pfcprule.PDR, fars []*pfcprule.FAR) error {
	// first add to temporary map to avoid erroring after msg is send
	tmpPDR := make(map[uint16]*pfcprule.PDR)
	for _, pdr := range pdrs {
		id, err := pdr.ID()
		if err != nil {
			return err
		}
		tmpPDR[id] = pdr
	}
	tmpFAR := make(map[uint32]*pfcprule.FAR)
	for _, far := range fars {
		id, err := far.ID()
		if err != nil {
			return err
		}
		tmpFAR[id] = far
	}
	ies := make([]*ie.IE, 0)
	ies = append(ies, s.association.Srv.NodeID())
	ies = append(ies, s.localFseid)
	for _, pdr := range pfcprule.NewCreatePDRs(pdrs) {
		ies = append(ies, pdr)
	}
	for _, far := range pfcprule.NewCreateFARs(fars) {
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
	s.AddFARs(tmpFAR)
	s.AddPDRs(tmpPDR)
	return nil
}
