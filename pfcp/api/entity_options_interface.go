// Copyright 2024 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT
package api

import "time"

type EntityOptionsInterface interface {
	MessageRetransmissionN1() int
	MessageRetransmissionT1() time.Duration
}
