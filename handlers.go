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
	res := message.NewAssociationSetupResponse(msg.Sequence(), entity.NodeID(), ie.NewCause(ie.CauseRequestAccepted), entity.RecoveryTimeStamp())
	return entity.ReplyTo(senderAddr, msg, res)
}
