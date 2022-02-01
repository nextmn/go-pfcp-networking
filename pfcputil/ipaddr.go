package pfcputil

import (
	"net"

	"github.com/wmnsk/go-pfcp/ie"
)

// Returns a NodeID IE from a string address
func CreateNodeID(id string) *ie.IE {
	ip := net.ParseIP(id)
	if ip == nil {
		// Node ID is a FQDN
		return ie.NewNodeID("", "", id)
	}
	if ip.To4() == nil {
		// Node ID is an IPv6 address
		return ie.NewNodeID("", id, "")
	}
	// Node ID is an IPv4 address
	return ie.NewNodeID(id, "", "")
}
