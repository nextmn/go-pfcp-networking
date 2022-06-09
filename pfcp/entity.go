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

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type handler = func(receivedMessage ReceivedMessage) error

type PFCPEntity struct {
	nodeID            *ie.IE
	recoveryTimeStamp *ie.IE
	handlers          map[pfcputil.MessageType]handler
	conn              *net.UDPConn
	connMu            sync.Mutex
	associationsMap   AssociationsMap
	// each session is associated with a specific PFCPAssociation
	// (can be changed with some requests)
	// UP function receives them from CP functions
	// CP function send them to UP functions
	sessionsMap SessionsMap
	kind        string // "CP" or "UP"
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

func newDefaultPFCPEntityHandlers() map[pfcputil.MessageType]handler {
	m := make(map[pfcputil.MessageType]handler)
	m[message.MsgTypeHeartbeatRequest] = handleHeartbeatRequest
	return m
}

func NewPFCPEntity(nodeID string, kind string) PFCPEntity {
	return PFCPEntity{
		nodeID:            ie.NewNodeIDHeuristic(nodeID),
		recoveryTimeStamp: nil,
		handlers:          newDefaultPFCPEntityHandlers(),
		conn:              nil,
		connMu:            sync.Mutex{},
		associationsMap:   NewAssociationsMap(),
		sessionsMap:       NewSessionsMap(),
		kind:              kind,
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

func (e *PFCPEntity) GetHandler(t pfcputil.MessageType) (h handler, err error) {
	if f, exists := e.handlers[t]; exists {
		return f, nil
	}
	return nil, fmt.Errorf("Received unexpected PFCP message type")
}

func (e *PFCPEntity) AddHandler(t pfcputil.MessageType, h handler) error {
	if e.RecoveryTimeStamp() != nil {
		return fmt.Errorf("Cannot add handler to already started PFCP Entity")
	}
	if !pfcputil.IsMessageTypeRequest(t) {
		return fmt.Errorf("Only request messages can have a handler")
	}
	e.handlers[t] = h
	return nil
}

func (e *PFCPEntity) AddHandlers(funcs map[pfcputil.MessageType]handler) error {
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

// Update an Association
func (e *PFCPEntity) UpdatePFCPAssociation(association api.PFCPAssociationInterface) error {
	return e.associationsMap.Update(association)
}

func (e *PFCPEntity) NewEstablishedPFCPAssociation(nodeID *ie.IE) (association api.PFCPAssociationInterface, err error) {
	peer, err := newPFCPPeerUP(e, nodeID)
	if err != nil {
		return nil, err
	}
	if e.RecoveryTimeStamp == nil {
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
