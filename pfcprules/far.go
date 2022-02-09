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
	f := make([]*ie.IE, 0)
	for _, far := range fars {
		f = append(f, far.NewCreateFAR())
	}
	return f
}

func NewFARs(fars []*ie.IE) (far []*FAR, err error, cause uint8, offendingIE uint16) {
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
			if err == io.ErrUnexpectedEOF {
				return nil, err, ie.CauseInvalidLength, ie.ForwardingParameters
			}
			if ie.NewApplyAction(aa).HasFORW() && err == ie.ErrIENotFound {
				return nil, err, ie.CauseConditionalIEMissing, ie.ForwardingParameters
			}
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
