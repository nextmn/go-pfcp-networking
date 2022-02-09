package pfcprule

import "github.com/wmnsk/go-pfcp/ie"

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

type PDRs []*PDR

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

func (pdr *PDR) ID() (uint16, error) {
	return pdr.id.PDRID()
}

func (pdr *PDR) PDI() ([]*ie.IE, error) {
	return pdr.pdi.PDI()
}
func (pdr *PDR) Precedence() (uint32, error) {
	return pdr.precedence.Precedence()
}

func (pdr *PDR) FARID() (uint32, error) {
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

func NewCreatePDRs(pdrs []*PDR) []*ie.IE {
	p := make([]*ie.IE, 0)
	for _, pdr := range pdrs {
		p = append(p, pdr.NewCreatePDR())
	}
	return p
}

func NewPDRs(pdrs []*ie.IE) ([]*PDR, error) {
	p := make([]*PDR, 0)
	for _, pdr := range pdrs {
		id, err := pdr.PDRID()
		if err != nil {
			return nil, err
		}
		pdi, err := pdr.PDI()
		if err != nil {
			return nil, err
		}
		precedence, err := pdr.Precedence()
		if err != nil {
			return nil, err
		}
		farid, err := pdr.FARID()
		if err != nil {
			return nil, err
		}

		// conditional IE
		var ohrIE *ie.IE
		if ohr, err := pdr.OuterHeaderRemoval(); err == nil {
			ohrIE = ie.NewOuterHeaderRemoval(ohr[0], ohr[1])
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
	return p, nil
}
