package pfcp_networking

import (
	"fmt"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func handleHeartbeatRequest(entity PFCPEntityInterface, senderAddr net.Addr, msg message.Message) error {
	log.Println("Received Heartbeat Request")
	res := message.NewHeartbeatResponse(msg.Sequence(), entity.RecoveryTimeStamp())
	return entity.ReplyTo(senderAddr, msg, res)
}

func handleAssociationSetupRequest(entity PFCPEntityInterface, senderAddr net.Addr, msg message.Message) error {
	log.Println("Received Association Setup Request")
	m, ok := msg.(*message.AssociationSetupRequest)
	if !ok {
		return fmt.Errorf("Issue with Association Setup Request")
	}
	peer, err := NewPFCPPeer(entity, m.NodeID)
	if err != nil {
		return err
	}
	association := NewAssociation(peer, entity)
	err = entity.CreatePFCPAssociation(&association)
	if err != nil {
		return err
	}
	switch {
	case msg == nil:
		return fmt.Errorf("msg is nil")
	case msg.Sequence == nil:
		return fmt.Errorf("msg.Sequence is nil")
	case entity == nil:
		return fmt.Errorf("entity is nil")
	case entity.NodeID() == nil:
		return fmt.Errorf("entity.NodeID() is nil")
	case entity.RecoveryTimeStamp() == nil:
		return fmt.Errorf("entity.RecoveryTimeStamp() is nil")
	}

	res := message.NewAssociationSetupResponse(msg.Sequence(), entity.NodeID(), ie.NewCause(ie.CauseRequestAccepted), entity.RecoveryTimeStamp())
	return entity.ReplyTo(senderAddr, msg, res)
}
