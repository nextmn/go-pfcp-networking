package pfcp_networking

import (
	"net"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPServerEntity struct {
	PFCPEntity
	associations   map[string]PFCPPeer
	muAssociations sync.Mutex
}

type serverHandler = func(serverEntity *PFCPServerEntity, senderAddr net.Addr, msg message.Message) error

func NewPFCPServerEntity(nodeID string) PFCPServerEntity {
	e := PFCPServerEntity{PFCPEntity: NewPFCPEntity(nodeID),
		associations:   make(map[string]PFCPPeer),
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
		return h(e, senderAddr, msg)
	}
	return e.AddHandler(t, f)
}

func (e *PFCPServerEntity) CreateAssociation(peer PFCPPeer) error {
	e.muAssociations.Lock()
	nid, err := peer.NodeID.NodeID()
	if err != nil {
		return err
	}
	e.associations[nid] = peer
	e.muAssociations.Unlock()
	return nil
}
