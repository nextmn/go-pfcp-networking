// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"net"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/nextmn/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/message"
)

type ReceivedMessage struct {
	message.Message
	SenderAddr net.Addr
	Entity     api.PFCPEntityInterface
}

func (receivedMessage *ReceivedMessage) NewResponse(responseMessage message.Message) (*OutcomingMessage, error) {
	if !pfcputil.IsMessageTypeRequest(receivedMessage.MessageType()) {
		return nil, fmt.Errorf("receivedMessage shall be a Request Message")
	}
	if !pfcputil.IsMessageTypeResponse(responseMessage.MessageType()) {
		return nil, fmt.Errorf("responseMessage shall be a Response Message")
	}
	if receivedMessage.Sequence() != responseMessage.Sequence() {
		return nil, fmt.Errorf("responseMessage shall have the same Sequence Number than receivedMessage")
	}
	return &OutcomingMessage{
		Message:     responseMessage,
		Destination: receivedMessage.SenderAddr,
	}, nil

}
