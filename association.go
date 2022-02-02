package pfcp_networking

import (
	"fmt"
	"time"
)

type PFCPAssociation struct {
	*PFCPPeer
	localEntity PFCPEntityInterface
}

func NewAssociation(peer *PFCPPeer, localEntity PFCPEntityInterface) PFCPAssociation {
	association := PFCPAssociation{PFCPPeer: peer, localEntity: localEntity}
	go association.heartMonitoring()
	return association
}

// Start monitoring heart of a PFCP Association
func (association *PFCPAssociation) heartMonitoring() error {
	defer association.Close()
	checkInterval := 30 * time.Second
	for {
		select {
		case <-time.After(checkInterval):
			alive, err := association.IsAlive()
			if !alive {
				return fmt.Errorf("PFCP Peer is dead")
			}
			if err != nil {
				return err
			}
		}
	}
}

// Close the association
func (association *PFCPAssociation) CloseAssociation() {
	association.Close()
	association.localEntity.RemovePFCPAssociation(association)
}
