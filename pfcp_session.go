package pfcp_networking

import (
	"fmt"
	"log"
	"sort"
	"sync"

	pfcprule "github.com/louisroyer/go-pfcp-networking/pfcprules"
	"github.com/wmnsk/go-pfcp/ie"
	"github.com/wmnsk/go-pfcp/message"
)

type PFCPSession struct {
	fseid     *ie.IE
	rseid     uint64
	pdr       map[uint16]*pfcprule.PDR
	far       map[uint32]*pfcprule.FAR
	sortedPDR pfcprule.PDRs
	mu        sync.Mutex
}

func NewPFCPSession(fseid *ie.IE, rseid uint64) PFCPSession {
	return PFCPSession{
		fseid:     fseid, // local F-SEID
		rseid:     rseid, // SEID present in FSEID ie send by remote peer
		pdr:       make(map[uint16]*pfcprule.PDR),
		far:       make(map[uint32]*pfcprule.FAR),
		sortedPDR: make(pfcprule.PDRs, 0),
		mu:        sync.Mutex{},
	}
}

func (s *PFCPSession) SEID() (uint64, error) {
	fseid, err := s.fseid.FSEID()
	if err != nil {
		return 0, err
	}
	return fseid.SEID, nil
}

func (s *PFCPSession) RSEID() uint64 {
	return s.rseid
}

func (s *PFCPSession) FSEID() *ie.IE {
	return s.fseid
}

func (s *PFCPSession) GetPDRs() pfcprule.PDRs {
	return s.sortedPDR
}

func (s *PFCPSession) GetFAR(farid uint32) (*pfcprule.FAR, error) {
	if far, ok := s.far[farid]; ok {
		return far, nil
	}
	return nil, fmt.Errorf("No far with id", farid)
}

func (s *PFCPSession) AddPDRs(pdrs map[uint16]*pfcprule.PDR) {
	s.mu.Lock()
	for id, pdr := range pdrs {
		s.pdr[id] = pdr
		s.sortedPDR = append(s.sortedPDR, pdr)
	}
	sort.Sort(s.sortedPDR)
	s.mu.Unlock()

}
func (s *PFCPSession) AddFARs(fars map[uint32]*pfcprule.FAR) {
	s.mu.Lock()
	for id, far := range fars {
		s.far[id] = far
	}
	s.mu.Unlock()
}

type RemotePFCPSession struct {
	PFCPSession
	association *PFCPAssociation
}

func NewRemotePFCPSession(fseid *ie.IE, association *PFCPAssociation) RemotePFCPSession {
	return RemotePFCPSession{
		PFCPSession: NewPFCPSession(fseid, 0), // Remote SEID is initialized when a Session Establishment Response is received
		association: association,
	}
}

func (s *RemotePFCPSession) Start(pdrs []*pfcprule.PDR, fars []*pfcprule.FAR) error {
	// first add to temporary map to avoid erroring after msg is send
	tmpPDR := make(map[uint16]*pfcprule.PDR)
	for _, pdr := range pdrs {
		id, err := pdr.ID()
		if err != nil {
			return err
		}
		tmpPDR[id] = pdr
	}
	tmpFAR := make(map[uint32]*pfcprule.FAR)
	for _, far := range fars {
		id, err := far.ID()
		if err != nil {
			return err
		}
		tmpFAR[id] = far
	}
	ies := make([]*ie.IE, 0)
	ies = append(ies, s.association.Srv.NodeID())
	ies = append(ies, s.fseid)
	for _, pdr := range pfcprule.NewCreatePDRs(pdrs) {
		ies = append(ies, pdr)
	}
	for _, far := range pfcprule.NewCreateFARs(fars) {
		ies = append(ies, far)
	}

	msg := message.NewSessionEstablishmentRequest(0, 0, 0, 0, 0, ies...)
	resp, err := s.association.Send(msg)
	if err != nil {
		return err
	}
	ser, ok := resp.(*message.SessionEstablishmentResponse)
	if !ok {
		log.Printf("got unexpected message: %s\n", resp.MessageTypeName())
	}
	fseid, err := ser.UPFSEID.FSEID()
	if err != nil {
		return err
	}
	s.rseid = fseid.SEID
	s.AddFARs(tmpFAR)
	s.AddPDRs(tmpPDR)
	return nil
}
