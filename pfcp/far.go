// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
	"github.com/wmnsk/go-pfcp/ie"
)

type FAR struct {
	id                   *ie.IE
	applyAction          *ie.IE
	forwardingParameters *ie.IE
}

func NewFAR(id *ie.IE, applyAction *ie.IE, forwardingParameters *ie.IE) *FAR {
	return &FAR{
		id:                   id,
		applyAction:          applyAction,
		forwardingParameters: forwardingParameters,
	}
}

func (far *FAR) ID() (api.FARID, error) {
	return far.id.FARID()
}

func (far *FAR) ApplyAction() *ie.IE {
	return far.applyAction
}

func (far *FAR) ForwardingParameters() *ie.IE {
	return far.forwardingParameters
}

func (far *FAR) NewCreateFAR() *ie.IE {
	ies := make([]*ie.IE, 0)
	ies = append(ies, far.id)
	ies = append(ies, far.applyAction)
	if far.forwardingParameters != nil && len(far.forwardingParameters) != 0 {
		ies = append(ies, far.forwardingParameters)
	}
	return ie.NewCreateFAR(ies...)
}
