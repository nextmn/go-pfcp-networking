package pfcp_networking

import (
	"fmt"
	"log"
	"net"

	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
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
	association := NewPFCPAssociation(peer)
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

func handleSessionEstablishmentRequest(entity PFCPEntityInterface, senderAddr net.Addr, msg message.Message) error {
	log.Println("Received Session Establishment Request")
	m, ok := msg.(*message.SessionEstablishmentRequest)
	if !ok {
		return fmt.Errorf("Issue with Session Establishment Request")
	}
	association, err := entity.GetPFCPAssociation(m.NodeID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fseid, err := m.CPFSEID.FSEID()
	if err != nil {
		return err
	}
	rseid := fseid.SEID
	if err != nil {
		return err
	}
	pdrs, err := pfcprule.NewPDRs(m.CreatePDR)
	if err != nil {
		return err
	}
	fars, err := pfcprule.NewFARs(m.CreateFAR)
	if err != nil {
		return err
	}
	session, err := association.CreateSession(entity.GetNextRemoteSessionID(), rseid, pdrs, fars)
	res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseRequestAccepted), session.FSEID())
	return entity.ReplyTo(senderAddr, msg, res)
}
