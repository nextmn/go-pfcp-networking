// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcprule

import (
	"io"

	"github.com/wmnsk/go-pfcp/ie"
)

type FAR struct {
	id                   *ie.IE
	applyAction          *ie.IE
	forwardingParameters *ie.IE
}

type FARID = uint32
type FARs []*FAR
type FARMap map[FARID]*FAR

func (far *FAR) ID() (FARID, error) {
	return far.id.FARID()
}

func (far *FAR) ApplyAction() *ie.IE {
	return far.applyAction
}

func NewFAR(id *ie.IE, applyAction *ie.IE, forwardingParameters *ie.IE) *FAR {
	return &FAR{
		id:                   id,
		applyAction:          applyAction,
		forwardingParameters: forwardingParameters,
	}
}

func (far *FAR) ForwardingParameters() *ie.IE {
	return far.forwardingParameters
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

func NewCreateFARs(fars FARMap) []*ie.IE {
	f := make([]*ie.IE, 0)
	for _, far := range fars {
		f = append(f, far.NewCreateFAR())
	}
	return f
}

func NewFARs(fars []*ie.IE) (far FARs, err error, cause uint8, offendingIE uint16) {
	f := make([]*FAR, 0)
	for _, far := range fars {
		id, err := far.FARID()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.FARID
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.FARID
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreateFAR
			}
		}
		aa, err := far.ApplyAction()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.ApplyAction
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.ApplyAction
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreateFAR
			}
		}
		fp, err := far.ForwardingParameters()
		// This IE shall be present when the Apply Action requests
		// the packets to be forwarded. It may be present otherwise.
		if err != nil {
			//XXX:  workaround for a free5gc-smf bug: Forwarding Parameters are missing sometimes
			fp = make([]*ie.IE, 0)
			//			if err == io.ErrUnexpectedEOF {
			//				return nil, err, ie.CauseInvalidLength, ie.ForwardingParameters
			//			}
			//			if ie.NewApplyAction(aa).HasFORW() && err == ie.ErrIENotFound {
			//				return nil, err, ie.CauseConditionalIEMissing, ie.ForwardingParameters
			//			}
		}

		f = append(f,
			NewFAR(
				ie.NewFARID(id),
				ie.NewApplyAction(aa),
				ie.NewForwardingParameters(fp...),
			))
	}
	return f, nil, 0, 0

}
