// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"time"

	"github.com/nextmn/go-pfcp-networking/pfcputil"
)

// entityOption is a function that updates fields of the given entityOptions parameter
// and returns another entityOption, which can be called to reverse this update.
type entityOption func(*entityOptions) entityOption

func OptionMessageRetransmissionT1(messageRetransmissionT1 time.Duration) entityOption {
	return func(eo *entityOptions) entityOption {
		if messageRetransmissionT1 < 0 {
			panic("messageRetransmissionT1 must be positive")
		}
		previous := eo.messageRetransmissionT1
		eo.messageRetransmissionT1 = messageRetransmissionT1
		return OptionMessageRetransmissionT1(previous)
	}
}

func OptionMessageRetransmissionN1(messageRetransmissionN1 int) entityOption {
	return func(eo *entityOptions) entityOption {
		if messageRetransmissionN1 < 0 {
			panic("messageRetransmissionN1 must be positive")
		}
		previous := eo.messageRetransmissionN1
		eo.messageRetransmissionN1 = messageRetransmissionN1
		return OptionMessageRetransmissionN1(previous)
	}
}

type entityOptions struct {
	messageRetransmissionT1 time.Duration
	messageRetransmissionN1 int
}

func (eo entityOptions) MessageRetransmissionT1() time.Duration {
	if eo.messageRetransmissionT1 == 0 {
		return pfcputil.MESSAGE_RETRANSMISSION_T1
	}
	return eo.messageRetransmissionT1
}

func (eo entityOptions) MessageRetransmissionN1() int {
	if eo.messageRetransmissionN1 == 0 {
		return pfcputil.MESSAGE_RETRANSMISSION_N1
	}
	return eo.messageRetransmissionN1
}
