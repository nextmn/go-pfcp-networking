// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/wmnsk/go-pfcp/ie"
)

type PDR struct {
	id                 *ie.IE
	pdi                *ie.IE
	precedence         *ie.IE
	farid              *ie.IE
	outerHeaderRemoval *ie.IE
}

func NewPDR(id *ie.IE, pdi *ie.IE, precedence *ie.IE, farid *ie.IE, outerHeaderRemoval *ie.IE) *PDR {
	return &PDR{
		id:                 id,
		pdi:                pdi,
		precedence:         precedence,
		farid:              farid,
		outerHeaderRemoval: outerHeaderRemoval,
	}
}

func (pdr *PDR) ID() (api.PDRID, error) {
	return pdr.id.PDRID()
}

func (pdr *PDR) PDI() ([]*ie.IE, error) {
	return pdr.pdi.PDI()
}
func (pdr *PDR) Precedence() (uint32, error) {
	return pdr.precedence.Precedence()
}

func (pdr *PDR) FARID() (api.FARID, error) {
	return pdr.farid.FARID()
}

func (pdr *PDR) OuterHeaderRemoval() *ie.IE {
	return pdr.outerHeaderRemoval
}

func (pdr *PDR) NewCreatePDR() *ie.IE {
	ies := make([]*ie.IE, 0)
	ies = append(ies, pdr.id)
	ies = append(ies, pdr.precedence)
	ies = append(ies, pdr.pdi)
	if pdr.outerHeaderRemoval != nil {
		ies = append(ies, pdr.outerHeaderRemoval)
	}
	if pdr.farid != nil {
		ies = append(ies, pdr.farid)
	}
	return ie.NewCreatePDR(ies...)
}
