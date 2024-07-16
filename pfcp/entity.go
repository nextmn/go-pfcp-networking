// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/nextmn/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntity struct {
	nodeID            *ie.IE
	recoveryTimeStamp *ie.IE
	handlers          map[pfcputil.MessageType]PFCPMessageHandler
	conn              *net.UDPConn
	connMu            sync.Mutex
	associationsMap   AssociationsMap
	// each session is associated with a specific PFCPAssociation
	// (can be changed with some requests)
	// UP function receives them from CP functions
	// CP function send them to UP functions
	sessionsMap api.SessionsMapInterface
	kind        string // "CP" or "UP"
	options     EntityOptions
}

func (e *PFCPEntity) Options() api.EntityOptionsInterface {
	return e.options
}

// Add an Established PFCP Session
func (e *PFCPEntity) AddEstablishedPFCPSession(session api.PFCPSessionInterface) error {
	return e.sessionsMap.Add(session)
}

func (e *PFCPEntity) GetPFCPSessions() []api.PFCPSessionInterface {
	return e.sessionsMap.GetPFCPSessions()
}

func (e *PFCPEntity) GetPFCPSession(localIP string, seid api.SEID) (api.PFCPSessionInterface, error) {
	return e.sessionsMap.GetPFCPSession(localIP, seid)
}

func (e *PFCPEntity) SendTo(msg []byte, dst net.Addr) error {
	e.connMu.Lock()
	defer e.connMu.Unlock()
	if _, err := e.conn.WriteTo(msg, dst); err != nil {
		return err
	}
	return nil
}

func (e *PFCPEntity) NodeID() *ie.IE {
	return e.nodeID
}
func (e *PFCPEntity) RecoveryTimeStamp() *ie.IE {
	return e.recoveryTimeStamp
}

func newDefaultPFCPEntityHandlers() map[pfcputil.MessageType]PFCPMessageHandler {
	m := make(map[pfcputil.MessageType]PFCPMessageHandler)
	m[message.MsgTypeHeartbeatRequest] = DefaultHeartbeatRequestHandler
	return m
}

func NewPFCPEntity(nodeID string, kind string, options EntityOptions) PFCPEntity {
	return PFCPEntity{
		nodeID:            ie.NewNodeIDHeuristic(nodeID),
		recoveryTimeStamp: nil,
		handlers:          newDefaultPFCPEntityHandlers(),
		conn:              nil,
		connMu:            sync.Mutex{},
		associationsMap:   NewAssociationsMap(),
		sessionsMap:       NewSessionsMap(),
		kind:              kind,
		options:           options,
	}
}

func (e *PFCPEntity) listen() error {
	e.recoveryTimeStamp = ie.NewRecoveryTimeStamp(time.Now())
	// TODO: if NodeID is a FQDN, we can expose multiple ip addresses
	ipAddr, err := e.NodeID().NodeID()
	if err != nil {
		return err
	}
	udpAddr := pfcputil.CreateUDPAddr(ipAddr, pfcputil.PFCP_PORT)
	laddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		return err
	}
	e.conn, err = net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	return nil
}

func (e *PFCPEntity) GetHandler(t pfcputil.MessageType) (h PFCPMessageHandler, err error) {
	if f, exists := e.handlers[t]; exists {
		return f, nil
	}
	return nil, fmt.Errorf("Received unexpected PFCP message type")
}

func (e *PFCPEntity) AddHandler(t pfcputil.MessageType, h PFCPMessageHandler) error {
	if e.RecoveryTimeStamp() != nil {
		return fmt.Errorf("Cannot add handler to already started PFCP Entity")
	}
	if !pfcputil.IsMessageTypeRequest(t) {
		return fmt.Errorf("Only request messages can have a handler")
	}
	e.handlers[t] = h
	return nil
}

func (e *PFCPEntity) AddHandlers(funcs map[pfcputil.MessageType]PFCPMessageHandler) error {
	if e.RecoveryTimeStamp() != nil {
		return fmt.Errorf("Cannot add handler to already started PFCP Entity")
	}
	for t, _ := range funcs {
		if !pfcputil.IsMessageTypeRequest(t) {
			return fmt.Errorf("Only request messages can have a handler")
		}
	}

	for t, h := range funcs {
		e.handlers[t] = h
	}
	return nil
}

// Remove an association from the association table
func (e *PFCPEntity) RemovePFCPAssociation(association api.PFCPAssociationInterface) error {
	return e.associationsMap.Remove(association)
}

// Returns an existing PFCP Association
func (e *PFCPEntity) GetPFCPAssociation(nid string) (association api.PFCPAssociationInterface, err error) {
	return e.associationsMap.Get(nid)
}

func (e *PFCPEntity) NewEstablishedPFCPAssociation(nodeID *ie.IE) (association api.PFCPAssociationInterface, err error) {
	peer, err := newPFCPPeerUP(e, nodeID)
	if err != nil {
		return nil, err
	}
	if e.RecoveryTimeStamp() == nil {
		return nil, fmt.Errorf("Local PFCP entity is not started")
	}
	nid, err := nodeID.NodeID()
	if err != nil {
		return nil, err
	}
	if !e.associationsMap.CheckNonExist(nid) {
		return nil, fmt.Errorf("Association already exists")
	}
	a, err := peer.NewEstablishedPFCPAssociation()
	if err != nil {
		return nil, err
	}
	if err := e.associationsMap.Add(a); err != nil {
		return nil, err
	}
	return a, nil

}

func (e *PFCPEntity) Start() error {
	if err := e.listen(); err != nil {
		return err
	}
	buf := make([]byte, pfcputil.DEFAULT_MTU) // TODO: get MTU of interface instead of using DEFAULT_MTU
	go func() error {
		for {
			n, addr, err := e.conn.ReadFrom(buf)
			if err != nil {
				return err
			}
			msg, err := message.Parse(buf[:n])
			if err != nil {
				// undecodable pfcp message
				continue
			}
			f, err := e.GetHandler(msg.MessageType())
			if err != nil {
				log.Println("No Handler for message of this type:", err)
				continue
			}
			err = f(ReceivedMessage{Message: msg, SenderAddr: addr, Entity: e})
			if err != nil {
				log.Println(err)
			}
		}
	}()
	return nil
}

func (e *PFCPEntity) IsUserPlane() bool {
	return e.kind == "UP"
}

func (e *PFCPEntity) IsControlPlane() bool {
	return e.kind == "CP"
}

func (e *PFCPEntity) PrintPFCPRules() {
	for _, session := range e.GetPFCPSessions() {
		localIPAddress, err := session.LocalIPAddress()
		if err != nil {
			log.Println(err)
			continue
		}
		localSEID, err := session.LocalSEID()
		if err != nil {
			log.Println(err)
			continue
		}
		remoteIPAddress, err := session.RemoteIPAddress()
		if err != nil {
			log.Println(err)
			continue
		}
		remoteSEID, err := session.RemoteSEID()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("PFCP Session: Local F-SEID [%s (%d)], Remote F-SEID [%s (%d)]\n",
			localIPAddress.String(), localSEID,
			remoteIPAddress.String(), remoteSEID)
		session.RLock()
		defer session.RUnlock()
		for _, pdrid := range session.GetSortedPDRIDs() {
			pdr, err := session.GetPDR(pdrid)
			if err != nil {
				log.Println(err)
				continue
			}
			precedence, err := pdr.Precedence()
			if err != nil {
				log.Println(err)
				continue
			}
			farid, err := pdr.FARID()
			if err != nil {
				log.Println(err)
				continue
			}
			pdicontent, err := pdr.PDI()
			if err != nil {
				log.Println(err)
				continue
			}
			far, err := session.GetFAR(farid)
			if err != nil {
				log.Println(err)
				continue
			}
			pdi := ie.NewPDI(pdicontent...)
			sourceInterfaceLabel := "Not defined"
			if sourceInterface, err := pdi.SourceInterface(); err == nil {
				switch sourceInterface {
				case ie.SrcInterfaceAccess:
					sourceInterfaceLabel = "Access"
				case ie.SrcInterfaceCore:
					sourceInterfaceLabel = "Core"
				case ie.SrcInterfaceSGiLANN6LAN:
					sourceInterfaceLabel = "SGi-LAN/N6-LAN"
				case ie.SrcInterfaceCPFunction:
					sourceInterfaceLabel = "CP Function"
				case ie.SrcInterface5GVNInternal:
					sourceInterfaceLabel = "5G VN Internal"
				}
			}
			ueIpAddressLabel := "Any"
			if ueipaddress, err := pdi.UEIPAddress(); err == nil {
				ueIpAddressIE := ie.NewUEIPAddress(ueipaddress.Flags, ueipaddress.IPv4Address.String(), ueipaddress.IPv6Address.String(), ueipaddress.IPv6PrefixDelegationBits, ueipaddress.IPv6PrefixLength)
				switch {
				case ueIpAddressIE.HasIPv4():
					ueIpAddressLabel = ueipaddress.IPv4Address.String()
				case ueIpAddressIE.HasIPv6():
					ueIpAddressLabel = ueipaddress.IPv6Address.String()
				}
			}
			fteidLabel := "Not defined"
			if fteid, err := pdi.FTEID(); err == nil {
				fteidIE := ie.NewFTEID(fteid.Flags, fteid.TEID, fteid.IPv4Address, fteid.IPv6Address, fteid.ChooseID)
				switch {
				case fteidIE.HasIPv4() && fteidIE.HasIPv6():
					fteidLabel = fmt.Sprintf("[%s/%s (%d)]", fteid.IPv4Address, fteid.IPv6Address, fteid.TEID)
				case fteidIE.HasIPv4():
					fteidLabel = fmt.Sprintf("[%s (%d)]", fteid.IPv4Address, fteid.TEID)
				case fteidIE.HasIPv6():
					fteidLabel = fmt.Sprintf("[%s (%d)]", fteid.IPv6Address, fteid.TEID)
				}
			}

			OuterHeaderRemovalLabel := "No"
			if ohrIe := pdr.OuterHeaderRemoval(); ohrIe != nil {
				if ohr, err := ohrIe.OuterHeaderRemovalDescription(); err == nil {
					if ohr == 0 || ohr == 1 || ohr == 6 {
						OuterHeaderRemovalLabel = "GTP"
					} else {
						OuterHeaderRemovalLabel = "Yes (but no GTP)"
					}
				}
			}

			SDFFilterLabel := ""
			if SDFFilter, err := pdi.SDFFilter(); err == nil {
				SDFFilterLabel = fmt.Sprintf("SDF Filter: %s", SDFFilter.FlowDescription)
			}

			ApplyActionLabel := "No"
			if ApplyActionIE := far.ApplyAction(); ApplyActionIE != nil {
				switch {
				case ApplyActionIE.HasDROP():
					ApplyActionLabel = "DROP"
				case ApplyActionIE.HasFORW():
					ApplyActionLabel = "FORW"
				default:
					ApplyActionLabel = "Other"
				}
			}

			ForwardingParametersIe := far.ForwardingParameters()
			OuterHeaderCreationLabel := "No"
			if ohc, err := ForwardingParametersIe.OuterHeaderCreation(); err == nil {
				ohcb, _ := ohc.Marshal()
				ohcIe := ie.New(ie.OuterHeaderCreation, ohcb)
				switch {
				case ohcIe.HasTEID() && ohcIe.HasIPv4():
					OuterHeaderCreationLabel = fmt.Sprintf("[%s (%d)]", ohc.IPv4Address.String(), ohc.TEID)
				case ohcIe.HasTEID() && ohcIe.HasIPv6():
					OuterHeaderCreationLabel = fmt.Sprintf("[%s (%d)]", ohc.IPv6Address.String(), ohc.TEID)
				default:
					OuterHeaderCreationLabel = "Other"
				}
			}

			DestinationInterfaceLabel := "Not defined"
			if destination, err := ForwardingParametersIe.DestinationInterface(); err == nil {
				switch destination {
				case ie.DstInterfaceAccess:
					DestinationInterfaceLabel = "Access"
				case ie.DstInterfaceCore:
					DestinationInterfaceLabel = "Core"
				case ie.DstInterfaceSGiLANN6LAN:
					DestinationInterfaceLabel = "SGi-LAN/N6-LAN"
				case ie.DstInterfaceCPFunction:
					DestinationInterfaceLabel = "CP Function"
				case ie.DstInterfaceLIFunction:
					DestinationInterfaceLabel = "LI Function"
				case ie.DstInterface5GVNInternal:
					DestinationInterfaceLabel = "5G VN Internal"
				}
			}

			log.Printf("  ↦ [PDR %d] (%d) Source interface: %s, OHR: %s, F-TEID: %s, UE IP: %s\n", pdrid, precedence, sourceInterfaceLabel, OuterHeaderRemovalLabel, fteidLabel, ueIpAddressLabel)
			if SDFFilterLabel != "" {
				log.Printf("    ↪ %s\n", SDFFilterLabel)
			}
			log.Printf("    ↪ [FAR %d] OHC: %s, ApplyAction: %s, Destination interface: %s\n", farid, OuterHeaderCreationLabel, ApplyActionLabel, DestinationInterfaceLabel)
		}
		log.Printf("\n")
	}
}
