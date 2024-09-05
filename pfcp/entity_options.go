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
	messageRetransmissionT1 time.Duration
	messageRetransmissionN1 int
}

// NewEntityOptions create a new EntityOptions with default settings.
func NewEntityOptions() *EntityOptions {
	return &EntityOptions{
		messageRetransmissionT1: pfcputil.MESSAGE_RETRANSMISSION_N1,
		messageRetransmissionN1: pfcputil.MESSAGE_RETRANSMISSION_N1,
	}
}

func (eo EntityOptions) MessageRetransmissionT1() time.Duration {
	return eo.messageRetransmissionT1
}

func (eo EntityOptions) SetMessageRetransmissionT1(messageRetransmissionT1 time.Duration) error {
	if messageRetransmissionT1 < 1*time.Microsecond {
		return fmt.Errorf("messageRetransmissionT1 must be strictly greater than zero.")
	}
	eo.messageRetransmissionT1 = messageRetransmissionT1
	return nil
}

func (eo EntityOptions) MessageRetransmissionN1() int {
	return eo.messageRetransmissionN1
}

func (eo EntityOptions) SetMessageRetransmissionN1(messageRetransmissionN1 int) error {
	if messageRetransmissionN1 < 0 {
		return fmt.Errorf("messageRetransmissionN1 must be greater than zero")
	}
	eo.messageRetransmissionN1 = messageRetransmissionN1
	return nil
}
