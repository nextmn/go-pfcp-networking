// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import "github.com/wmnsk/go-pfcp/ie"

type FARID = uint32

type FARInterface interface {
	ID() (FARID, error)
	ApplyAction() *ie.IE
	ForwardingParameters() (*ie.IE, error)
	NewCreateFAR() *ie.IE
	Update(FARUpdateInterface) error
}
