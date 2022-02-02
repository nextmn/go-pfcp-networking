package pfcp_networking

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntityInterface interface {
	NodeID() *ie.IE
	RecoveryTimeStamp() *ie.IE
	CreatePFCPAssociation(association *PFCPAssociation) error
	RemovePFCPAssociation(association *PFCPAssociation) error
	ReplyTo(ipAddress net.Addr, requestMessage message.Message, responseMessage message.Message) error
}

func (entity *PFCPEntity) ReplyTo(ipAddress net.Addr, requestMessage message.Message, responseMessage message.Message) error {
	if !pfcputil.IsMessageTypeRequest(requestMessage.MessageType()) {
		return fmt.Errorf("requestMessage shall be a Request Message")
	}
	if !pfcputil.IsMessageTypeResponse(responseMessage.MessageType()) {
		return fmt.Errorf("responseMessage shall be a Response Message")
	}
	if requestMessage.Sequence() != responseMessage.Sequence() {
		return fmt.Errorf("responseMessage shall have the same Sequence Number than requestMessage")
	}
	//XXX: message.Message interface does not implement Marshal()
	b := make([]byte, responseMessage.MarshalLen())
	if err := responseMessage.MarshalTo(b); err != nil {
		return err
	}

	entity.mu.Lock()
	if _, err := entity.conn.WriteTo(b, ipAddress); err != nil {
		return err
	}
	entity.mu.Unlock()
	return nil
}

type handler = func(entity PFCPEntityInterface, senderAddr net.Addr, msg message.Message) error

type PFCPEntity struct {
	nodeID            *ie.IE
	recoveryTimeStamp *ie.IE
	handlers          map[pfcputil.MessageType]handler
	conn              *net.UDPConn
	mu                sync.Mutex
	iface             PFCPEntityInterface
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

func NewPFCPEntity(nodeID string) PFCPEntity {
	return PFCPEntity{
		nodeID:            pfcputil.CreateNodeID(nodeID),
		recoveryTimeStamp: nil,
		handlers:          newDefaultPFCPEntityHandlers(),
		conn:              nil,
		mu:                sync.Mutex{},
		iface:             nil,
	}
}

func (e *PFCPEntity) Start() error {
	if e.iface == nil {
		return fmt.Errorf("PFCPEntity is incorrectly initialized")
	}
	e.recoveryTimeStamp = ie.NewRecoveryTimeStamp(time.Now())
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
			if f, exists := e.handlers[msg.MessageType()]; exists {
				err := f(e.iface, addr, msg)
				if err != nil {
					log.Println(err)
				}
			} else {
				log.Println("Received unexpected PFCP message type")
			}
		}
	}()
	return nil
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
