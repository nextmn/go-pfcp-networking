// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcputil

import "time"

const (
	// TODO: detect MTU of interface automatically
	DEFAULT_MTU = 1500

	// The UDP Destination Port number for a Request message shall be 8805.
	// It is the registered port number for PFCP.
	PFCP_PORT = 8805

	// When sending a Request message, the sending PFCP entity shall start a timer T1.
	// The sending entity shall consider that the Request message has been lost if
	// a corresponding Response message has not been received before the T1 timer expires.
	// If so, the sending entity shall retransmit the Request message,
	// if the total number of retry attempts is less than N1 times.
	// The setting of the T1 timer and N1 counter is implementation specific.
	MESSAGE_RETRANSMISSION_T1 = time.Millisecond * 500
	MESSAGE_RETRANSMISSION_N1 = 3
)
