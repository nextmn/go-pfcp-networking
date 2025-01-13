// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"github.com/nextmn/go-pfcp-networking/pfcp/api"
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

func (far *FAR) ForwardingParameters() (*ie.IE, error) {
	// This IE shall be present when the Apply Action requests
	// the packets to be forwarded. It may be present otherwise.
	if far.forwardingParameters == nil {
		return nil, fmt.Errorf("No forwarding parameters ie")
	}
	return far.forwardingParameters, nil

}

func (far *FAR) NewCreateFAR() *ie.IE {
	ies := make([]*ie.IE, 0)
	ies = append(ies, far.id)
	ies = append(ies, far.applyAction)
	if far.forwardingParameters != nil {
		ies = append(ies, far.forwardingParameters)
	}
	return ie.NewCreateFAR(ies...)
}

func (far *FAR) Update(farUpdate api.FARUpdateInterface) error {
	// Check FARID
	farUpdateId, err := farUpdate.ID()
	if err != nil {
		return err
	}
	farId, err := far.ID()
	if err != nil {
		return err
	}
	if farId != farUpdateId {
		return fmt.Errorf("Wrong FAR ID")
	}

	// Update ApplyAction
	if aa := farUpdate.ApplyAction(); aa != nil {
		far.applyAction = aa
	}

	// Update Forwarding Parameters
	ForwParam := make([]*ie.IE, 0)
	if fp := farUpdate.UpdateForwardingParameters(); fp != nil {
		if di, err := fp.DestinationInterface(); err == nil {
			ForwParam = append(ForwParam, ie.NewDestinationInterface(di))
		} else if diorg, err := far.forwardingParameters.DestinationInterface(); err == nil {
			ForwParam = append(ForwParam, ie.NewDestinationInterface(diorg))
		}
		if ni, err := fp.NetworkInstance(); err == nil {
			ForwParam = append(ForwParam, ie.NewNetworkInstance(ni))
		} else if niorg, err := far.forwardingParameters.NetworkInstance(); err == nil {
			ForwParam = append(ForwParam, ie.NewNetworkInstance(niorg))
		}
		if ohc, err := fp.OuterHeaderCreation(); err == nil {
			ForwParam = append(ForwParam, ie.NewOuterHeaderCreation(ohc.OuterHeaderCreationDescription, ohc.TEID, ohc.IPv4Address.String(), ohc.IPv6Address.String(), ohc.PortNumber, ohc.CTag, ohc.STag))
		} else if ohcorg, err := far.forwardingParameters.OuterHeaderCreation(); err == nil {
			ForwParam = append(ForwParam, ie.NewOuterHeaderCreation(ohcorg.OuterHeaderCreationDescription, ohcorg.TEID, ohcorg.IPv4Address.String(), ohcorg.IPv6Address.String(), ohcorg.PortNumber, ohcorg.CTag, ohcorg.STag))
		}
		far.forwardingParameters = ie.NewForwardingParameters(ForwParam...)
	}

	return nil
}
