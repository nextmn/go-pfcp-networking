// Copyright 2024 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package pfcp_networking

import (
	"fmt"
	"net"

	"github.com/nextmn/go-pfcp-networking/pfcputil"
)

type PFCPConn struct {
	net.UDPConn
}

func ListenPFCP(network string, laddr *net.IPAddr) (*PFCPConn, error) {
	switch network {
	case "udp", "udp4", "udp6":
	default:
		return nil, fmt.Errorf("unknown network")
	}
	udpAddr := pfcputil.CreateUDPAddr(laddr.String(), pfcputil.PFCP_PORT)
	ludpaddr, err := net.ResolveUDPAddr(network, udpAddr)
	if err != nil {
		return nil, err
	}
	if conn, err := net.ListenUDP(network, ludpaddr); err == nil {
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
