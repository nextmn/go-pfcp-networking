package pfcp_networking

import (
	"log"
	"sync"

	"github.com/wmnsk/go-pfcp/message"
)

type PFCPServerEntity struct {
	PFCPEntity
	associations   map[string]*PFCPAssociation
	muAssociations sync.Mutex
}

func NewPFCPServerEntity(nodeID string) PFCPServerEntity {
	e := PFCPServerEntity{PFCPEntity: NewPFCPEntity(nodeID),
		associations:   make(map[string]*PFCPAssociation),
		muAssociations: sync.Mutex{},
	}
	e.iface = &e
	err := e.initDefaultHandlers()
	if err != nil {
		log.Println(err)
	}
	return e
}

func (e *PFCPServerEntity) initDefaultHandlers() error {
	return e.AddHandler(message.MsgTypeAssociationSetupRequest, handleAssociationSetupRequest)
}

// Add an association to the association table
func (e PFCPServerEntity) CreatePFCPAssociation(association *PFCPAssociation) error {
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
func (e PFCPServerEntity) RemovePFCPAssociation(association *PFCPAssociation) error {
	nid, err := association.NodeID.NodeID()
	if err != nil {
		return err
	}
	e.muAssociations.Lock()
	delete(e.associations, nid)
	e.muAssociations.Unlock()
	return nil
}
