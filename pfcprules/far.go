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

func (far *FAR) ForwardingParameters() ([]*ie.IE, error) {
	return far.forwardingParameters.ForwardingParameters()
}

func (far *FAR) NewCreateFAR() *ie.IE {
	return ie.NewCreatePDR(
		far.id,
		far.applyAction,
		far.forwardingParameters,
	)
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
