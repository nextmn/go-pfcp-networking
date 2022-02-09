package pfcp_networking

import (
	"fmt"
	"io"
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
	// CP F-SEID is a mandatory IE, but if it is missing or malformed, we are not able to send response
	fseid, err := m.CPFSEID.FSEID()
	if err != nil {
		return err
	}
	rseid := fseid.SEID
	if err != nil {
		return err
	}
	// check NodeID is present, or send cause(Mandatory IE missing)
	if m.NodeID == nil {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseMandatoryIEMissing), ie.NewOffendingIE(ie.NodeID))
		return entity.ReplyTo(senderAddr, msg, res)
	}
	nid, err := m.NodeID.NodeID()
	// check NodeID is well-formed, or send cause(Mandatory IE incorrect)
	if err != nil {
		cause := ie.CauseMandatoryIEIncorrect
		if err == io.ErrUnexpectedEOF {
			cause = ie.CauseInvalidLength
		}
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(cause), ie.NewOffendingIE(ie.NodeID))
		return entity.ReplyTo(senderAddr, msg, res)
	}

	association, err := entity.GetPFCPAssociation(nid)
	// check Association is established, or send cause(No established PFCP Association)
	if err != nil {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseNoEstablishedPFCPAssociation))
		return entity.ReplyTo(senderAddr, msg, res)
	}

	// CreatePDR is a Mandatory IE
	if m.CreatePDR == nil || len(m.CreatePDR) == 0 {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseMandatoryIEMissing), ie.NewOffendingIE(ie.CreatePDR))
		return entity.ReplyTo(senderAddr, msg, res)
	}

	// CreateFAR is a Mandatory IE
	if m.CreateFAR == nil || len(m.CreateFAR) == 0 {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseMandatoryIEMissing), ie.NewOffendingIE(ie.CreateFAR))
		return entity.ReplyTo(senderAddr, msg, res)
	}
	pdrs, err, cause, offendingie := pfcprule.NewPDRs(m.CreatePDR)
	if err != nil {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(cause), ie.NewOffendingIE(offendingie))
		return entity.ReplyTo(senderAddr, msg, res)
	}
	fars, err, cause, offendingie := pfcprule.NewFARs(m.CreateFAR)
	if err != nil {
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(cause), ie.NewOffendingIE(offendingie))
		return entity.ReplyTo(senderAddr, msg, res)
	}
	session, err := association.CreateSession(entity.GetNextRemoteSessionID(), rseid, pdrs, fars)
	if err != nil {
		// Send cause(Rule creation/modification failure)
		res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseRuleCreationModificationFailure))
		return entity.ReplyTo(senderAddr, msg, res)

	}
	res := message.NewSessionEstablishmentResponse(0, 0, rseid, msg.Sequence(), 0, entity.NodeID(), ie.NewCause(ie.CauseRequestAccepted), session.FSEID())
	return entity.ReplyTo(senderAddr, msg, res)
}
