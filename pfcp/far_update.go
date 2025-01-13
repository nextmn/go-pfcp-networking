// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"github.com/nextmn/go-pfcp-networking/pfcp/api"

	"github.com/wmnsk/go-pfcp/ie"
)

type FARUpdate struct {
	id                         *ie.IE
	applyAction                *ie.IE
	updateForwardingParameters *ie.IE
}

func NewFARUpdate(id *ie.IE, applyAction *ie.IE, updateForwardingParameters *ie.IE) *FARUpdate {
	return &FARUpdate{
		id:                         id,
		applyAction:                applyAction,
		updateForwardingParameters: updateForwardingParameters,
	}
}

func (far *FARUpdate) ID() (api.FARID, error) {
	return far.id.FARID()
}

func (far *FARUpdate) ApplyAction() *ie.IE {
	return far.applyAction
}

func (far *FARUpdate) UpdateForwardingParameters() *ie.IE {
	return far.updateForwardingParameters
}

func (far *FARUpdate) NewUpdateFAR() *ie.IE {
	ies := make([]*ie.IE, 0)
	ies = append(ies, far.id)
	if far.applyAction != nil {
		ies = append(ies, far.applyAction)
	}
	if far.updateForwardingParameters != nil {
		ies = append(ies, far.updateForwardingParameters)
	}
	return ie.NewUpdateFAR(ies...)
}
