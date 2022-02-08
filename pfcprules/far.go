package pfcprule

import "github.com/wmnsk/go-pfcp/ie"

type FAR struct {
	id                   *ie.IE
	applyAction          *ie.IE
	forwardingParameters *ie.IE
}

func (far *FAR) ID() (uint32, error) {
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

func (far *FAR) ForwardingParameters() ([]*ie.IE, error) {
	return far.forwardingParameters.ForwardingParameters()
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

func NewCreateFARs(fars []*FAR) []*ie.IE {
	f := make([]*ie.IE, len(fars))
	for i, far := range fars {
		f[i] = far.NewCreateFAR()
	}
	return f
}

func NewFARs(fars []*ie.IE) (far []*FAR, err error) {
	f := make([]*FAR, len(fars))
	for _, far := range fars {
		id, err := far.FARID()
		if err != nil {
			return nil, err
		}
		aa, err := far.ApplyAction()
		if err != nil {
			return nil, err
		}
		fp, err := far.ForwardingParameters()
		if err != nil {
			return nil, err
		}

		f = append(f,
			&FAR{
				ie.NewFARID(id),
				ie.NewApplyAction(aa),
				ie.NewForwardingParameters(fp...),
			})
	}
	return f, nil

}
