package pfcprule

import "github.com/wmnsk/go-pfcp/ie"

type PDR struct {
	id                 *ie.IE
	pdi                *ie.IE
	precedence         *ie.IE
	farid              *ie.IE
	outerHeaderRemoval *ie.IE
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
	return ie.NewCreatePDR(
		pdr.id,
		pdr.precedence,
		pdr.pdi,
		pdr.outerHeaderRemoval,
		pdr.farid,
	)
}

func NewCreatePDRs(pdrs []*PDR) []*ie.IE {
	p := make([]*ie.IE, len(pdrs))
	for i, pdr := range pdrs {
		p[i] = pdr.NewCreatePDR()
	}
	return p
}

func NewPDRs(pdrs []*ie.IE) ([]*PDR, error) {
	p := make([]*PDR, len(pdrs))
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
		ohr, err := pdr.OuterHeaderRemoval()
		if err != nil {
			return nil, err
		}

		p = append(p,
			&PDR{
				ie.NewPDRID(id),
				ie.NewPDI(pdi...),
				ie.NewPrecedence(precedence),
				ie.NewFARID(farid),
				ie.NewOuterHeaderRemoval(ohr[0], ohr[1]),
			})
	}
	return p, nil
}
