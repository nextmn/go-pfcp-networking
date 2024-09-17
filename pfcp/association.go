// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"net"
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/sirupsen/logrus"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPAssociation struct {
	api.PFCPPeerInterface                // connection to remote peer
	isSetup               bool           // true when session is already set-up
	sessionIDPool         *SessionIDPool // used to generate SEIDs for this association
}

// Create a new PFCPAssociation, this association is already set-up
func newEstablishedPFCPAssociation(peer api.PFCPPeerInterface) (api.PFCPAssociationInterface, error) {
	association := PFCPAssociation{
		PFCPPeerInterface: peer,
		isSetup:           false,
		sessionIDPool:     NewSessionIDPool(),
	}
	err := association.SetupInitiatedByCP()
	if err != nil {
		return nil, err
	}
	return &association, nil
}

// Get next available SEID for this PFCPAssociation.
// SEID are not globally unique, F-SEID are globally unique.
// F-SEID consist of IPv4 and/or IPv6 address(es) of the peer
// plus the SEID. So as long as SEID are unique per peer (i.e. per PFCPAssociation),
// everything should be okay.
func (association *PFCPAssociation) GetNextSEID() api.SEID {
	return association.sessionIDPool.GetNext()
}

// Setup a PFCPAssociation with the PFCP Association Setup by the CP Function Procedure
// If the LocalEntity is a CP function, a PFCP Association Setup Request is sent,
// if the LocalEntity is a UP function, we assume this method is called
// because we received a Association Setup Request.
//
// PFCP Association Setup Initiated by the UP Function is NOT supported (yet). PR are welcome.
//
// See 129.244 v16.0.1, section 6.2.6.1:
// The setup of a PFCP association may be initiated by the CP function (see clause 6.2.6.2) or the UP function (see
// clause 6.2.6.3).
// The CP function and the UP function shall support the PFCP Association Setup initiated by the CP function. The CP
// function and the UP function may additionally support the PFCP Association Setup initiated by the UP function.
func (association *PFCPAssociation) SetupInitiatedByCP() error {
	if association.isSetup {
		return fmt.Errorf("Association is already set up")
	}
	switch {
	case association.LocalEntity().IsUserPlane():
		association.isSetup = true
		go association.heartMonitoring()
		return nil
	case association.LocalEntity().IsControlPlane():
		sar := message.NewAssociationSetupRequest(0, association.LocalEntity().NodeID(), association.LocalEntity().RecoveryTimeStamp())
		resp, err := association.Send(sar)
		if err != nil {
			return err
		}
		asres, ok := resp.(*message.AssociationSetupResponse)
		if !ok {
			logrus.WithFields(logrus.Fields{"message-type": resp.MessageTypeName()}).Debug("Got unexpected message")
		}
		cause, err := asres.Cause.Cause()
		if err != nil {
			// TODO: send missing ie message
			return err
		}
		if cause == ie.CauseRequestAccepted {
			association.isSetup = true
			go association.heartMonitoring()
			return nil
		}
		return fmt.Errorf("Association setup request rejected")
	default:
		return fmt.Errorf("Local PFCP entity is not a UP function, neither a CP function.")
	}
}

// Start monitoring heart of a PFCP Association
func (association *PFCPAssociation) heartMonitoring() error {
	defer association.Close()
	checkInterval := 30 * time.Second
	for {
		select {
		case <-time.After(checkInterval):
			alive, err := association.IsAlive()
			if !alive {
				return fmt.Errorf("PFCP Peer is dead")
			}
			if err != nil {
				return err
			}
		}
	}
}

// Generate a local FSEID IE for the session (to be created) identified by a given SEID
func (association *PFCPAssociation) getFSEID(seid api.SEID) (*ie.IE, error) {
	ieNodeID := association.LocalEntity().NodeID()
	nodeID, err := ieNodeID.NodeID()
	if err != nil {
		return nil, err
	}
	var localFseid *ie.IE
	switch ieNodeID.Payload[0] {
	case ie.NodeIDIPv4Address:
		ip4, err := net.ResolveIPAddr("ip4", nodeID)
		if err != nil {
			return nil, err
		}
		localFseid, err = NewFSEID(seid, ip4, nil)
		if err != nil {
			return nil, err
		}
	case ie.NodeIDIPv6Address:
		ip6, err := net.ResolveIPAddr("ip6", nodeID)
		if err != nil {
			return nil, err
		}
		localFseid, err = NewFSEID(seid, nil, ip6)
		if err != nil {
			return nil, err
		}
	case ie.NodeIDFQDN:
		ip4, err4 := net.ResolveIPAddr("ip4", nodeID)
		ip6, err6 := net.ResolveIPAddr("ip6", nodeID)
		if err4 != nil && err6 != nil {
			return nil, fmt.Errorf("Cannot resolve NodeID")
		}
		switch {
		case err4 == nil && err6 == nil:
			localFseid = ie.NewFSEID(seid, ip4.IP.To4(), ip6.IP.To16())
		case err4 == nil && err6 != nil:
			localFseid = ie.NewFSEID(seid, ip4.IP.To4(), nil)
		case err4 != nil && err6 == nil:
			localFseid = ie.NewFSEID(seid, nil, ip6.IP.To16())
		case err4 != nil && err6 != nil:
			return nil, fmt.Errorf("Cannot resolve NodeID")
		}
	}
	return localFseid, nil
}

// remoteFseid can be nil if caller is at CP function side
func (association *PFCPAssociation) CreateSession(remoteFseid *ie.IE, pdrs api.PDRMapInterface, fars api.FARMapInterface) (session api.PFCPSessionInterface, err error) {
	// Generation of the F-SEID
	localSEID := association.GetNextSEID()
	localFseid, err := association.getFSEID(localSEID)
	if err != nil {
		return nil, err
	}
	// Establishment of a PFCP Session if CP / Creation if UP
	s, err := newEstablishedPFCPSession(association, localFseid, remoteFseid, pdrs, fars)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Safe function to create FSEID
func NewFSEID(seid api.SEID, v4, v6 *net.IPAddr) (*ie.IE, error) {
	if v4 == nil && v6 == nil {
		return nil, fmt.Errorf("Cannot create FSEID with no IP Address")
	}
	var ip4, ip6 net.IP
	if v4 != nil {
		ip4 = v4.IP.To4()
	}
	if v6 != nil {
		ip6 = v6.IP.To16()
	}
	return ie.NewFSEID(seid, ip4, ip6), nil
}
