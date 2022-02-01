package pfcp_networking

import (
	"github.com/wmnsk/go-pfcp/ie"
)

type PFCPPeer struct {
	NodeID *ie.IE
}

func NewPFCPPeer(nodeID *ie.IE) PFCPPeer {
	return PFCPPeer{
		NodeID: nodeID,
	}
}
