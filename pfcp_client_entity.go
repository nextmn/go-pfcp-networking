package pfcp_networking

import (
	"fmt"
	"log"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPClientEntity struct {
	PFCPEntity
	associations   map[string]*PFCPAssociation
	muAssociations sync.Mutex
}

func NewPFCPClientEntity(nodeID string) *PFCPClientEntity {
	e := PFCPClientEntity{PFCPEntity: NewPFCPEntity(nodeID),
		associations:   make(map[string]*PFCPAssociation),
		muAssociations: sync.Mutex{},
	}
	return &e
}

// Add an association to the association table
func (e *PFCPClientEntity) CreatePFCPAssociation(association *PFCPAssociation) error {
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
func (e *PFCPClientEntity) RemovePFCPAssociation(association *PFCPAssociation) error {
	nid, err := association.NodeID.NodeID()
	if err != nil {
		return err
	}
	e.muAssociations.Lock()
	delete(e.associations, nid)
	e.muAssociations.Unlock()
	return nil
}

// Returns an existing PFCP Association
func (e *PFCPClientEntity) GetPFCPAssociation(nodeID *ie.IE) (association *PFCPAssociation, err error) {
	nid, err := e.NodeID().NodeID()
	if err != nil {
		return nil, err
	}
	if a, exists := e.associations[nid]; exists {
		return a, nil
	}
	return nil, fmt.Errorf("Association does not exist.")
}

// Create a PFCP Association, by sending a PFCP Association Setup Request
func (e *PFCPClientEntity) NewPFCPAssociation(peer *PFCPPeer) (association *PFCPAssociation, err error) {
	if e.RecoveryTimeStamp == nil {
		return nil, fmt.Errorf("Local PFCP entity is not started")
	}
	nid, err := peer.NodeID.NodeID()
	if err != nil {
		return nil, err
	}
	if _, exists := e.associations[nid]; exists {
		return nil, fmt.Errorf("Only one association shall be setup between given pair of CP and UP functions.")
	}
	sar := message.NewAssociationSetupRequest(0, e.NodeID(), e.RecoveryTimeStamp())
	resp, err := peer.Send(sar)
	if err != nil {
		return nil, err
	}
	asres, ok := resp.(*message.AssociationSetupResponse)
	if !ok {
		log.Printf("got unexpected message: %s\n", resp.MessageTypeName())
	}
	cause, err := asres.Cause.Cause()
	if err != nil {
		// TODO: send missing ie message
		return nil, err
	}
	if cause == ie.CauseRequestAccepted {
		a := NewPFCPAssociation(peer, e)
		e.CreatePFCPAssociation(&a)
		return &a, nil
	}
	return nil, fmt.Errorf("Associaton setup request rejected")
}

func (e *PFCPClientEntity) Start() error {
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
