package pfcp_networking

import (
	"fmt"
	"log"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPServerEntity struct {
	PFCPEntity
	associations   map[string]*PFCPAssociation
	muAssociations sync.Mutex
}

func NewPFCPServerEntity(nodeID string) *PFCPServerEntity {
	e := PFCPServerEntity{PFCPEntity: NewPFCPEntity(nodeID),
		associations:   make(map[string]*PFCPAssociation),
		muAssociations: sync.Mutex{},
	}
	err := e.initDefaultHandlers()
	if err != nil {
		log.Println(err)
	}
	return &e
}

func (e *PFCPServerEntity) initDefaultHandlers() error {
	if err := e.AddHandler(message.MsgTypeAssociationSetupRequest, handleAssociationSetupRequest); err != nil {
		return err
	}
	if err := e.AddHandler(message.MsgTypeSessionEstablishmentRequest, handleSessionEstablishmentRequest); err != nil {
		return err
	}
	return nil
}

// Add an association to the association table
func (e *PFCPServerEntity) CreatePFCPAssociation(association *PFCPAssociation) error {
	nid, err := association.NodeID.NodeID()
	if err != nil {
		return err
	}
	e.muAssociations.Lock()
	e.associations[nid] = association
	e.muAssociations.Unlock()
	return nil
}

// Returns an existing PFCP Association
func (e *PFCPServerEntity) GetPFCPAssociation(nodeID *ie.IE) (association *PFCPAssociation, err error) {
	nid, err := e.NodeID().NodeID()
	if err != nil {
		return nil, err
	}
	if a, exists := e.associations[nid]; exists {
		return a, nil
	}
	return nil, fmt.Errorf("Association does not exist.")
}

// Remove an association from the association table
func (e *PFCPServerEntity) RemovePFCPAssociation(association *PFCPAssociation) error {
	nid, err := association.NodeID.NodeID()
	if err != nil {
		return err
	}
	e.muAssociations.Lock()
	delete(e.associations, nid)
	e.muAssociations.Unlock()
	return nil
}

func (e *PFCPServerEntity) Start() error {
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
				log.Println(err)
			}
			err = f(e, addr, msg)
			if err != nil {
				log.Println(err)
			}
		}
	}()
	return nil
}

func (e *PFCPServerEntity) GetPFCPSessions() []*PFCPSession {
	sessions := make([]*PFCPSession, 0)
	for _, a := range e.associations {
		as := a.GetSessions()
		for _, s := range as {
			sessions = append(sessions, s)
		}
	}
	return sessions
}
