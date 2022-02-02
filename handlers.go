package pfcp_networking

import (
	"fmt"
	"log"
	"net"

	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

func handleHeartbeatRequest(entity *PFCPEntity, senderAddr net.Addr, msg message.Message) error {
	res := message.NewHeartbeatResponse(msg.Sequence(), entity.RecoveryTimeStamp)
	return entity.ReplyTo(senderAddr, msg, res)
}

func handleAssociationSetupRequest(srv *PFCPServerEntity, senderAddr net.Addr, msg message.Message) error {
	m, ok := msg.(*message.AssociationSetupRequest)
	if !ok {
		return fmt.Errorf("Issue with Association Setup Request")
	}
	peer, err := NewPFCPPeer(&srv.PFCPEntity, m.NodeID)
	if err != nil {
		return err
	}
	association := NewAssociation(peer, srv)
	err = srv.CreatePFCPAssociation(&association)
	if err != nil {
		return err
	}
	switch {
	case msg == nil:
		log.Println("msg is nil")
		fallthrough
	case srv == nil:
		log.Println("srv is nil")
		fallthrough
	case srv.NodeID == nil:
		log.Println("srv.NodeID is nil")
		fallthrough
	case srv.RecoveryTimeStamp == nil:
		log.Println("srv.RecoveryTimeStamp is nil")
	}
	res := message.NewAssociationSetupResponse(msg.Sequence(), srv.NodeID, ie.NewCause(ie.CauseRequestAccepted), srv.RecoveryTimeStamp)
	return srv.ReplyTo(senderAddr, msg, res)
}
