// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"net"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
	"github.com/louisroyer/go-pfcp-networking/pfcputil"
	"github.com/wmnsk/go-pfcp/message"
)

type ReceivedMessage struct {
	message.Message
	SenderAddr net.Addr
	Entity     api.PFCPEntityInterface
}

func (receivedMessage *ReceivedMessage) ReplyTo(responseMessage message.Message) error {
	if !pfcputil.IsMessageTypeRequest(receivedMessage.MessageType()) {
		return fmt.Errorf("receivedMessage shall be a Request Message")
	}
	if !pfcputil.IsMessageTypeResponse(responseMessage.MessageType()) {
		return fmt.Errorf("responseMessage shall be a Response Message")
	}
	if receivedMessage.Sequence() != responseMessage.Sequence() {
		return fmt.Errorf("responseMessage shall have the same Sequence Number than receivedMessage")
	}
	//XXX: message.Message interface does not implement Marshal()
	b := make([]byte, responseMessage.MarshalLen())
	if err := responseMessage.MarshalTo(b); err != nil {
		return err
	}
	if err := receivedMessage.Entity.SendTo(b, receivedMessage.SenderAddr); err != nil {
		return err
	}
	return nil
}
