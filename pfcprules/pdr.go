// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcprule

import (
	"io"

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

type PDRID = uint16
type PDRs []*PDR
type PDRMap map[PDRID]*PDR

func (pdrs PDRs) Less(i, j int) bool {
	// element with highest precedence (lowest value in Precedence IE) should be sorted first
	pi, err := pdrs[i].precedence.Precedence()
	if err != nil {
		return false
	}
	pj, err := pdrs[j].precedence.Precedence()
	if err != nil {
		return true
	}
	return pi < pj
}

func (pdrs PDRs) Len() int {
	return len(pdrs)
}
func (pdrs PDRs) Swap(i, j int) {
	pdrs[i], pdrs[j] = pdrs[j], pdrs[i]
}

func (pdr *PDR) ID() (PDRID, error) {
	return pdr.id.PDRID()
}

func (pdr *PDR) PDI() ([]*ie.IE, error) {
	return pdr.pdi.PDI()
}
func (pdr *PDR) Precedence() (uint32, error) {
	return pdr.precedence.Precedence()
}

func (pdr *PDR) FARID() (FARID, error) {
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

func NewCreatePDRs(pdrs PDRMap) []*ie.IE {
	p := make([]*ie.IE, 0)
	for _, pdr := range pdrs {
		p = append(p, pdr.NewCreatePDR())
	}
	return p
}

func NewPDRs(pdrs []*ie.IE) (p []*PDR, err error, cause uint8, offendingIE uint16) {
	for _, pdr := range pdrs {
		id, err := pdr.PDRID()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.PDRID
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.PDRID
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreatePDR
			}
		}
		pdi, err := pdr.PDI()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.PDI
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.PDI
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreatePDR
			}
		}
		precedence, err := pdr.Precedence()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.Precedence
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.Precedence
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreatePDR
			}
		}
		farid, err := pdr.FARID()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, err, ie.CauseInvalidLength, ie.FARID
			case ie.ErrIENotFound:
				return nil, err, ie.CauseMandatoryIEMissing, ie.FARID
			default:
				return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreatePDR
			}
		}

		// conditional IE
		var ohrIE *ie.IE
		ohr, err := pdr.OuterHeaderRemoval()
		if err == nil {
			// ohr can be 1 byte lenght with old format
			if len(ohr) == 1 {
				ohr = append(ohr, 0)
			}
			ohrIE = ie.NewOuterHeaderRemoval(ohr[0], ohr[1])
		} else if err == io.ErrUnexpectedEOF {
			return nil, err, ie.CauseInvalidLength, ie.OuterHeaderRemoval
		}

		p = append(p,
			NewPDR(
				ie.NewPDRID(id),
				ie.NewPDI(pdi...),
				ie.NewPrecedence(precedence),
				ie.NewFARID(farid),
				ohrIE,
			))
	}
	return p, nil, 0, 0
}
