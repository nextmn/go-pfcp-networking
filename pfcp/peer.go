// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/nextmn/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type messageChan chan []byte

// A PFCPPeer is a remote PFCPEntity
type PFCPPeer struct {
	nodeID  *ie.IE
	srv     api.PFCPEntityInterface
	conn    *net.UDPConn
	udpAddr *net.UDPAddr
	seq     uint32
	seqMu   sync.Mutex
	queue   map[uint32]messageChan
	queueMu sync.Mutex
	stop    bool
	kind    string
}

func (peer *PFCPPeer) NewEstablishedPFCPAssociation() (api.PFCPAssociationInterface, error) {
	return newEstablishedPFCPAssociation(peer)
}

func (peer *PFCPPeer) LocalEntity() api.PFCPEntityInterface {
	return peer.srv
}

func (peer *PFCPPeer) NodeID() *ie.IE {
	return peer.nodeID
}
func newPFCPPeer(srv api.PFCPEntityInterface, nodeID *ie.IE, kind string) (peer *PFCPPeer, err error) {
	ipAddr, err := nodeID.NodeID()
	if err != nil {
		return nil, err
	}
	udpAddr := pfcputil.CreateUDPAddr(ipAddr, pfcputil.PFCP_PORT)
	raddr, err := net.ResolveUDPAddr("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	c, err := net.Dial("udp", udpAddr)
	if err != nil {
		return nil, err
	}
	c.Close()
	laddr := c.LocalAddr().(*net.UDPAddr)
	conn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}
	p := PFCPPeer{
		srv:     srv,
		nodeID:  nodeID,
		conn:    conn,
		udpAddr: raddr,
		seq:     1,
		seqMu:   sync.Mutex{},
		queue:   make(map[uint32]messageChan),
		queueMu: sync.Mutex{},
		stop:    false,
		kind:    kind,
	}
	// Read incomming messages
	p.start()
	return &p, nil
}

func newPFCPPeerUP(srv api.PFCPEntityInterface, nodeID *ie.IE) (peer *PFCPPeer, err error) {
	return newPFCPPeer(srv, nodeID, "UP")
}
func newPFCPPeerCP(srv api.PFCPEntityInterface, nodeID *ie.IE) (peer *PFCPPeer, err error) {
	return newPFCPPeer(srv, nodeID, "CP")
}

func (peer *PFCPPeer) IsUserPlane() bool {
	return peer.kind == "UP"
}

func (peer *PFCPPeer) IsControlPlane() bool {
	return peer.kind == "CP"
}

func (peer *PFCPPeer) start() {
	go func(e *PFCPPeer) {
		e.startLoop()
	}(peer)
}
func (peer *PFCPPeer) startLoop() {
	for peer.IsRunning() {
		peer.loopUnwrapped()
	}
}

func (peer *PFCPPeer) loopUnwrapped() {
	b := make([]byte, pfcputil.DEFAULT_MTU) // TODO: detect MTU for interface instead of using DEFAULT_MTU
	n, _, err := peer.conn.ReadFromUDP(b)
	if err != nil {
		// socket has been closed
		return
	}
	// Processing of message in a new thread to avoid blocking
	go func(msgArray []byte, size int, e *PFCPPeer) {
		msg, err := message.ParseHeader(msgArray[:size])
		if err != nil {
			return
		}
		sn := msg.SequenceNumber

		e.queueMu.Lock()
		defer e.queueMu.Unlock()
		ch, exists := e.queue[sn]
		if exists {
			ch <- msgArray[:size]
		}
	}(b, n, peer)
}

func (peer *PFCPPeer) IsRunning() bool {
	return !peer.stop
}

// Close connection of PFCPPeer
func (peer *PFCPPeer) Close() error {
	// if already stopped, for whatever reason, we exit
	if peer.stop {
		return nil
	}
	// setting stop state and closing connection
	peer.stop = true
	err := peer.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Get next sequence number available for this PFCPPeer
// Sequence murber shall be unique for each oustanding
// message sourced from the same IP/UDP endpoint.
// Since we use exactly 1 IP/UDP endpoint per peer to send Requests,
// our sequence numbers are also unique per peer.
func (peer *PFCPPeer) getNextSequenceNumber() uint32 {
	peer.seqMu.Lock()
	defer peer.seqMu.Unlock()
	s := peer.seq
	peer.seq += 1
	return s
}

// Add a message to queue. Response will be send to channel ch messageChan
func (peer *PFCPPeer) addToQueue(sn uint32, ch messageChan) {
	peer.queueMu.Lock()
	defer peer.queueMu.Unlock()
	peer.queue[sn] = ch
}

// Remove a message from queue (used when a response is received, or when timeout is reached)
func (peer *PFCPPeer) deleteFromQueue(sn uint32) {
	peer.queueMu.Lock()
	defer peer.queueMu.Unlock()
	close(peer.queue[sn])
	delete(peer.queue, sn)
}

// Send a PFCP message
func (peer *PFCPPeer) Send(msg message.Message) (m message.Message, err error) {

	//XXX: cannot use `h, err := msg.(*message.Header)` because Header does not implement MessageTypeName()
	msgb := make([]byte, msg.MarshalLen())
	err = msg.MarshalTo(msgb)
	if err != nil {
		return nil, err
	}
	h, err := message.ParseHeader(msgb)
	if err != nil {
		return nil, err
	}

	if !pfcputil.IsMessageTypeRequest(h.MessageType()) {
		return nil, fmt.Errorf("Unexpected outcomming PFCP message type")
	}
	sn := peer.getNextSequenceNumber()
	h.SetSequenceNumber(sn)
	b, err := h.Marshal()
	if err != nil {
		return nil, err
	}

	ch := make(messageChan)
	peer.addToQueue(sn, ch)
	defer peer.deleteFromQueue(sn)

	_, err = peer.conn.WriteToUDP(b, peer.udpAddr)
	if err != nil {
		return nil, fmt.Errorf("Error on write: %s\n", err)
	}

	for i := 0; i < pfcputil.MESSAGE_RETRANSMISSION_N1; i++ {
		select {
		case r := <-ch:
			msg, err := message.Parse(r)
			if err != nil {
				return nil, fmt.Errorf("Unexpected incomming packet")
			}
			if !pfcputil.IsMessageTypeResponse(msg.MessageType()) {
				return nil, fmt.Errorf("Unexpected incomming PFCP message type")
			}
			return msg, nil
		case <-time.After(pfcputil.MESSAGE_RETRANSMISSION_T1):
			// retry
			_, err = peer.conn.WriteToUDP(b, peer.udpAddr)
			if err != nil {
				return nil, fmt.Errorf("Error on write: %s\n", err)
			}
		}
	}
	return nil, fmt.Errorf("Unsuccessfull transfer of Request message")
}

// Send an Heartbeat request, return true if the PFCP peer is alive.
func (peer *PFCPPeer) IsAlive() (res bool, err error) {
	if peer.LocalEntity().RecoveryTimeStamp() == nil {
		return false, fmt.Errorf("Local PFCP Entity is not yet started.")
	}
	hreq := message.NewHeartbeatRequest(
		0,
		peer.LocalEntity().RecoveryTimeStamp(),
		nil)

	_, err = peer.Send(hreq)
	if err != nil {
		return false, err
	}
	return true, nil
}
