// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcputil

import (
	"github.com/wmnsk/go-pfcp/message"
)

type MessageType = uint8

// Returns true when message is a PFCP Request
func IsMessageTypeRequest(msgType MessageType) bool {
	// Not using
	// `return !isMessagetypeResponse(msgType)`
	// because new message Types may be defined
	// in the futur
	switch msgType {
	case message.MsgTypeHeartbeatRequest:
		fallthrough
	case message.MsgTypePFDManagementRequest:
		fallthrough
	case message.MsgTypeAssociationSetupRequest:
		fallthrough
	case message.MsgTypeAssociationUpdateRequest:
		fallthrough
	case message.MsgTypeAssociationReleaseRequest:
		fallthrough
	case message.MsgTypeNodeReportRequest:
		fallthrough
	case message.MsgTypeSessionSetDeletionRequest:
		fallthrough
	case message.MsgTypeSessionEstablishmentRequest:
		fallthrough
	case message.MsgTypeSessionModificationRequest:
		fallthrough
	case message.MsgTypeSessionDeletionRequest:
		fallthrough
	case message.MsgTypeSessionReportRequest:
		return true
	default:
		return false
	}
}

// Returns true when message is a PFCP Response
func IsMessageTypeResponse(msgType MessageType) bool {
	// Not using
	// `return !isMessagetypeRequest(msgType)`
	// because new message Types may be defined
	// in the futur
	switch msgType {
	case message.MsgTypeHeartbeatResponse:
		fallthrough
	case message.MsgTypePFDManagementResponse:
		fallthrough
	case message.MsgTypeAssociationSetupResponse:
		fallthrough
	case message.MsgTypeAssociationUpdateResponse:
		fallthrough
	case message.MsgTypeAssociationReleaseResponse:
		fallthrough
	case message.MsgTypeVersionNotSupportedResponse:
		fallthrough
	case message.MsgTypeNodeReportResponse:
		fallthrough
	case message.MsgTypeSessionSetDeletionResponse:
		fallthrough
	case message.MsgTypeSessionEstablishmentResponse:
		fallthrough
	case message.MsgTypeSessionModificationResponse:
		fallthrough
	case message.MsgTypeSessionDeletionResponse:
		fallthrough
	case message.MsgTypeSessionReportResponse:
		return true
	default:
		return false
	}
}
