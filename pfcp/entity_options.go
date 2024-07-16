// Copyright 2024 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcputil"
)

type EntityOptions struct {
	messageRetransmissionT1 *time.Duration
	messageRetransmissionN1 *int
}

func NewEntityOptions(messageRetransmissionT1 *time.Duration, messageRetransmissionN1 *int) (*EntityOptions, error) {
	if (messageRetransmissionT1 != nil) && (*messageRetransmissionT1 < 1*time.Microsecond) {
		return nil, fmt.Errorf("messageRetransmissionT1 must be strictly greater than zero.")
	}
	if (messageRetransmissionN1 != nil) && (*messageRetransmissionN1 < 0) {
		return nil, fmt.Errorf("messageRetransmissionN1 must be greater than zero")
	}
	return &EntityOptions{
		messageRetransmissionT1: messageRetransmissionT1,
		messageRetransmissionN1: messageRetransmissionN1,
	}, nil
}

func (eo EntityOptions) MessageRetransmissionT1() time.Duration {
	if eo.messageRetransmissionT1 != nil {
		return *eo.messageRetransmissionT1
	} else {
		return pfcputil.MESSAGE_RETRANSMISSION_T1
	}
}

func (eo EntityOptions) MessageRetransmissionN1() int {
	if eo.messageRetransmissionN1 != nil {
		return *eo.messageRetransmissionN1
	} else {
		return pfcputil.MESSAGE_RETRANSMISSION_N1
	}
}
