// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package pfcp_networking

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/nextmn/go-pfcp-networking/pfcputil"
)

type PFCPConn struct {
	net.UDPConn
}

func ListenPFCP(network string, laddr netip.Addr) (*PFCPConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, fmt.Errorf("unknown network")
	}
	udpaddr := net.UDPAddrFromAddrPort(netip.AddrPortFrom(laddr, pfcputil.PFCP_PORT))
	if conn, err := net.ListenUDP(network, udpaddr); err == nil {
		return &PFCPConn{
			UDPConn: *conn,
		}, nil
	} else {
		return nil, err
	}
}

func (conn *PFCPConn) Write(m *OutcomingMessage) error {
	//XXX: message.Message interface does not implement Marshal()
	b := make([]byte, m.MarshalLen())
	if err := m.MarshalTo(b); err != nil {
		return err
	}
	if _, err := conn.WriteTo(b, m.Destination); err != nil {
		return err
	}
	return nil
}
