package pfcp_networking

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPEntityInterface interface {
	RemovePFCPAssociation(association *PFCPAssociation) error
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

type handler = func(entity *PFCPEntity, senderAddr net.Addr, msg message.Message) error

type PFCPEntity struct {
	NodeID            *ie.IE
	RecoveryTimeStamp *ie.IE
	handlers          map[pfcputil.MessageType]handler
	conn              *net.UDPConn
	mu                sync.Mutex
}

func newDefaultPFCPEntityHandlers() map[pfcputil.MessageType]handler {
	m := make(map[pfcputil.MessageType]handler)
	m[message.MsgTypeHeartbeatRequest] = handleHeartbeatRequest
	return m
}

func NewPFCPEntity(nodeID string) PFCPEntity {
	return PFCPEntity{
		NodeID:            pfcputil.CreateNodeID(nodeID),
		RecoveryTimeStamp: nil,
		handlers:          newDefaultPFCPEntityHandlers(),
		conn:              nil,
		mu:                sync.Mutex{},
	}
}

func (e *PFCPEntity) Start() error {
	e.RecoveryTimeStamp = ie.NewRecoveryTimeStamp(time.Now())
	ipAddr, err := e.NodeID.NodeID()
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
				f(e, addr, msg)
			}
		}
	}()
	return nil
}

func (e *PFCPEntity) AddHandler(t pfcputil.MessageType, h handler) error {
	if e.RecoveryTimeStamp != nil {
		return fmt.Errorf("Cannot add handler to already started PFCP Entity")
	}
	if !pfcputil.IsMessageTypeRequest(t) {
		return fmt.Errorf("Only request messages can have a handler")
	}
	e.handlers[t] = h
	return nil
}

func (e *PFCPEntity) AddHandlers(funcs map[pfcputil.MessageType]handler) error {
	if e.RecoveryTimeStamp != nil {
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
