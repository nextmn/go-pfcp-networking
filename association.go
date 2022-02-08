package pfcp_networking

import (
	"fmt"
	"net"
	"sync"
	"time"

	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPAssociation struct {
	*PFCPPeer
	localEntity    PFCPEntityInterface
	sessions       map[uint64]*PFCPSession
	remoteSessions map[uint64]*RemotePFCPSession
	mu             sync.Mutex
}

func NewPFCPAssociation(peer *PFCPPeer, localEntity PFCPEntityInterface) PFCPAssociation {
	association := PFCPAssociation{
		PFCPPeer:       peer,
		localEntity:    localEntity,
		sessions:       make(map[uint64]*PFCPSession),
		remoteSessions: make(map[uint64]*RemotePFCPSession),
	}
	go association.heartMonitoring()
	return association
}

func (association *PFCPAssociation) getNextRemoteSessionID() uint64 {
	fmt.Println("Calling get Next Remote Session ID")
	id := association.localEntity.GetNextRemoteSessionID()
	fmt.Println("Next Remote Session ID is ", id)
	return id
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

func (association *PFCPAssociation) GetSessions() map[uint64]*PFCPSession {
	return association.sessions
}

func (association *PFCPAssociation) getFSEID() (*ie.IE, error) {
	seid := association.getNextRemoteSessionID()
	ieNodeID := association.Srv.NodeID()
	nodeID, err := ieNodeID.NodeID()
	if err != nil {
		return nil, err
	}
	var localFseid *ie.IE
	switch ieNodeID.Payload[0] {
	case ie.NodeIDIPv4Address:
		ip4, err := net.ResolveIPAddr("ip4", nodeID)
		if err != nil {
			return nil, err
		}
		localFseid, err = NewFSEID(seid, ip4, nil)
		if err != nil {
			return nil, err
		}
	case ie.NodeIDIPv6Address:
		ip6, err := net.ResolveIPAddr("ip6", nodeID)
		if err != nil {
			return nil, err
		}
		localFseid, err = NewFSEID(seid, nil, ip6)
		if err != nil {
			return nil, err
		}
	case ie.NodeIDFQDN:
		ip4, err4 := net.ResolveIPAddr("ip4", nodeID)
		ip6, err6 := net.ResolveIPAddr("ip6", nodeID)
		if err4 != nil && err6 != nil {
			return nil, fmt.Errorf("Cannot resolve NodeID")
		}
		localFseid = ie.NewFSEID(seid, ip4.IP.To4(), ip6.IP.To16())
	}
	return localFseid, nil
}
func (association *PFCPAssociation) CreateSession(rseid uint64, pdrs []*pfcprule.PDR, fars []*pfcprule.FAR) (session *PFCPSession, err error) {
	localFseid, err := association.getFSEID()
	if err != nil {
		return nil, err
	}
	s := NewPFCPSession(localFseid, rseid)
	tmpPDR := make(map[uint16]*pfcprule.PDR)
	for _, pdr := range pdrs {
		id, err := pdr.ID()
		if err != nil {
			return nil, err
		}
		tmpPDR[id] = pdr
	}
	tmpFAR := make(map[uint32]*pfcprule.FAR)
	for _, far := range fars {
		id, err := far.ID()
		if err != nil {
			return nil, err
		}
		tmpFAR[id] = far
	}
	s.AddFARs(tmpFAR)
	s.AddPDRs(tmpPDR)
	association.mu.Lock()
	association.sessions[rseid] = &s
	association.mu.Unlock()
	return &s, nil
}

func (association *PFCPAssociation) NewPFCPSession(pdrs []*pfcprule.PDR, fars []*pfcprule.FAR) (session *RemotePFCPSession, err error) {
	localFseid, err := association.getFSEID()
	if err != nil {
		return nil, err
	}
	s := NewRemotePFCPSession(localFseid, association)
	s.Start(pdrs, fars)
	rseid := s.RSEID()
	association.mu.Lock()
	association.remoteSessions[rseid] = &s
	association.mu.Unlock()
	return &s, nil
}

func NewFSEID(seid uint64, v4, v6 *net.IPAddr) (*ie.IE, error) {
	if v4 == nil && v6 == nil {
		return nil, fmt.Errorf("Cannot create FSEID with no IP Address")
	}
	var ip4, ip6 net.IP
	if v4 != nil {
		ip4 = v4.IP.To4()
	}
	if v6 != nil {
		ip6 = v6.IP.To16()
	}
	return ie.NewFSEID(seid, ip4, ip6), nil
}
