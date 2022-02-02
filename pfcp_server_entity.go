package pfcp_networking

import (
	"fmt"
	"net"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPServerEntity struct {
	PFCPEntity
	associations   map[string]*PFCPAssociation
	muAssociations sync.Mutex
}

type serverHandler = func(serverEntity *PFCPServerEntity, senderAddr net.Addr, msg message.Message) error

func NewPFCPServerEntity(nodeID string) PFCPServerEntity {
	e := PFCPServerEntity{PFCPEntity: NewPFCPEntity(nodeID),
		associations:   make(map[string]*PFCPAssociation),
		muAssociations: sync.Mutex{},
	}
	e.initDefaultHandlers()
	return e
}

func (e *PFCPServerEntity) initDefaultHandlers() error {
	return e.AddServerHandler(message.MsgTypeAssociationSetupRequest, handleAssociationSetupRequest)
}

func (e *PFCPServerEntity) AddServerHandler(t pfcputil.MessageType, h serverHandler) error {
	f := func(entity *PFCPEntity, senderAddr net.Addr, msg message.Message) error {
		if e == nil {
			return fmt.Errorf("PFCPServerEntity is nil")
		}
		if e.RecoveryTimeStamp == nil {
			return fmt.Errorf("RecoveryTimestamp is nil")
		}
		return h(e, senderAddr, msg)
	}
	return e.AddHandler(t, f)
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
