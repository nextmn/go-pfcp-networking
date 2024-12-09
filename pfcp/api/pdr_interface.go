// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import "github.com/wmnsk/go-pfcp/ie"

type PDRID = uint16
type PDRInterface interface {
	ID() (PDRID, error)

	PDI() ([]*ie.IE, error)
	Precedence() (uint32, error)

	FARID() (FARID, error)

	OuterHeaderRemoval() *ie.IE

	SourceInterface() (uint8, error)
	FTEID() (*ie.FTEIDFields, error)
	UEIPAddress() (*ie.UEIPAddressFields, error)

	NewCreatePDR() *ie.IE
	NewUpdatePDR() *ie.IE
}
