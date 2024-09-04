// Copyright 2024 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"net"

	"github.com/wmnsk/go-pfcp/message"
)

type OutcomingMessage struct {
	message.Message
	Destination net.Addr
}
