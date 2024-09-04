// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/nextmn/go-pfcp-networking/pfcputil"
	"github.com/sirupsen/logrus"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntity struct {
	nodeID            *ie.IE
	listenAddr        string
	recoveryTimeStamp *ie.IE
	handlers          map[pfcputil.MessageType]PFCPMessageHandler
	associationsMap   AssociationsMap
	// each session is associated with a specific PFCPAssociation
	// (can be changed with some requests)
	// UP function receives them from CP functions
	// CP function send them to UP functions
	sessionsMap api.SessionsMapInterface
	kind        string // "CP" or "UP"
	options     api.EntityOptionsInterface

	mu        sync.Mutex
	pfcpConns []*onceClosePfcpConn
	closeFunc context.CancelFunc
}

// onceClosePfcpConn wraps a PfcpConn, protecting it from multiple Close calls.
type onceClosePfcpConn struct {
	*PFCPConn
	once     sync.Once
	closeErr error
}

func (oc *onceClosePfcpConn) Close() error {
	oc.once.Do(oc.close)
	return oc.closeErr
}

func (oc *onceClosePfcpConn) close() {
	oc.closeErr = oc.PFCPConn.Close()
}

func (e *PFCPEntity) registerPfcpConn(conn *onceClosePfcpConn) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.pfcpConns == nil {
		e.pfcpConns = make([]*onceClosePfcpConn, 1)
	}
	e.pfcpConns = append(e.pfcpConns, conn)
}

func (e *PFCPEntity) closePfcpConn() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.pfcpConns == nil {
		return nil
	}
	for _, v := range e.pfcpConns {
		if err := v.Close(); err != nil {
			return err
		}
	}
	e.pfcpConns = nil
	return nil
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

func NewPFCPEntity(nodeID string, listenAddr string, kind string, handlers map[pfcputil.MessageType]PFCPMessageHandler, options api.EntityOptionsInterface) PFCPEntity {
	if handlers == nil {
		handlers = newDefaultPFCPEntityHandlers()
	}
	return PFCPEntity{
		nodeID:            ie.NewNodeIDHeuristic(nodeID),
		listenAddr:        listenAddr,
		recoveryTimeStamp: ie.NewRecoveryTimeStamp(time.Now()),
		handlers:          handlers,
		associationsMap:   NewAssociationsMap(),
		sessionsMap:       NewSessionsMap(),
		kind:              kind,
		options:           options,
		pfcpConns:         nil,
	}
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

// Listen PFCP and run the entity with the provided context.
// Always return a non-nil error.
func (e *PFCPEntity) ListenAndServe() error {
	return e.ListenAndServeContext(context.Background())
}

// Listen PFCP and run the entity with the provided context.
// Always return a non-nil error.
func (e *PFCPEntity) ListenAndServeContext(ctx context.Context) error {
	// TODO: listen on both ipv4 and ipv6
	ipaddr, err := net.ResolveIPAddr("ip", e.listenAddr)
	if err != nil {
		return err
	}
	if conn, err := ListenPFCP("udp", ipaddr); err != nil {
		return err
	} else {
		e.Serve(ctx, conn)
		return fmt.Errorf("Server closed")
	}
}

// Run the entity with the provided context.
// Always return a non-nil error and close the PFCPConn.
func (e *PFCPEntity) Serve(ctx context.Context, conn *PFCPConn) error {
	if conn == nil {
		return fmt.Errorf("Conn is nil")
	}
	newconn := &onceClosePfcpConn{PFCPConn: conn}
	e.registerPfcpConn(newconn)
	serveCtx, cancel := context.WithCancel(ctx)
	e.closeFunc = cancel
	defer newconn.Close()
	for {
		select {
		case <-serveCtx.Done():
			// Stop signal received
			return serveCtx.Err()
		default:
			buf := make([]byte, pfcputil.DEFAULT_MTU) // TODO: get MTU of interface instead of using DEFAULT_MTU
			if n, addr, err := conn.ReadFrom(buf); err == nil {
				go func(ctx context.Context, buffer []byte, sender net.Addr) {
					msg, err := message.Parse(buffer)
					if err != nil {
						// undecodable pfcp message
						return
					}
					f, err := e.GetHandler(msg.MessageType())
					if err != nil {
						logrus.WithFields(logrus.Fields{"message-type": msg.MessageType}).WithError(err).Debug("No Handler for message of this type")
						return
					}
					if resp, err := f(ctx, ReceivedMessage{Message: msg, SenderAddr: addr, Entity: e}); err != nil {
						logrus.WithError(err).Debug("Handler raised an error")
					} else {
						select {
						case <-ctx.Done():
							return
						default:
							conn.Write(resp)
						}
					}
				}(serveCtx, buf[:n], addr)
			}
		}
	}
}

// Close stop the server and closes active PFCP connection.
func (e *PFCPEntity) Close() error {
	e.closeFunc()
	e.closePfcpConn()
	return nil
}

func (e *PFCPEntity) IsUserPlane() bool {
	return e.kind == "UP"
}

func (e *PFCPEntity) IsControlPlane() bool {
	return e.kind == "CP"
}

func (e *PFCPEntity) LogPFCPRules() {
	for _, session := range e.GetPFCPSessions() {
		localIPAddress, err := session.LocalIPAddress()
		if err != nil {
			logrus.WithError(err).Debug("Could not get local IP Address")
			continue
		}
		localSEID, err := session.LocalSEID()
		if err != nil {
			logrus.WithError(err).Debug("Could not get local SEID")
			continue
		}
		remoteIPAddress, err := session.RemoteIPAddress()
		if err != nil {
			logrus.WithError(err).Debug("Could not get remote IP Address")
			continue
		}
		remoteSEID, err := session.RemoteSEID()
		if err != nil {
			logrus.WithError(err).Debug("Could not get remote SEID")
			continue
		}

		session.RLock()
		defer session.RUnlock()
		for _, pdrid := range session.GetSortedPDRIDs() {
			pdr, err := session.GetPDR(pdrid)
			if err != nil {
				logrus.WithError(err).Debug("Could not get PDR")
				continue
			}
			precedence, err := pdr.Precedence()
			if err != nil {
				logrus.WithError(err).Debug("Could not get Precedence")
				continue
			}
			farid, err := pdr.FARID()
			if err != nil {
				logrus.WithError(err).Debug("Could not get FAR ID")
				continue
			}
			pdicontent, err := pdr.PDI()
			if err != nil {
				logrus.WithError(err).Debug("Could not get PDI")
				continue
			}
			far, err := session.GetFAR(farid)
			if err != nil {
				logrus.WithError(err).Debug("Could not get FAR")
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

			logrus.WithFields(logrus.Fields{
				"session/local-fseid":      localIPAddress.String(),
				"session/local-seid":       localSEID,
				"session/remote-fseid":     remoteIPAddress.String(),
				"session/remote-seid":      remoteSEID,
				"pdr/id":                   pdrid,
				"pdr/precedence":           precedence,
				"pdr/source-iface":         sourceInterfaceLabel,
				"pdr/outer-header-removal": OuterHeaderRemovalLabel,
				"pdr/fteid":                fteidLabel,
				"pdr/ue-ip-addr":           ueIpAddressLabel,
				"pdr/sdf-filter":           SDFFilterLabel,
				"far/id":                   farid,
				"far/ohc":                  OuterHeaderCreationLabel,
				"far/apply-action":         ApplyActionLabel,
				"far/destination-iface":    DestinationInterfaceLabel,
			}).Info("PDR")

		}
	}
}
