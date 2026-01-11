// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"io"
	"sync"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"

	"github.com/wmnsk/go-pfcp/ie"
)

type farmapUpdateInternal = map[api.FARID]api.FARUpdateInterface

type FARMapUpdate struct {
	farmap farmapUpdateInternal
	mu     sync.RWMutex
}

func NewFARMapUpdate(fars []*ie.IE) (*FARMapUpdate, error, uint8, uint16) {
	f := FARMapUpdate{
		farmap: make(farmapUpdateInternal),
		mu:     sync.RWMutex{},
	}
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
		var ieaa *ie.IE = nil
		aa, err := far.ApplyAction()
		if err == nil {
			ieaa = ie.NewApplyAction(aa...)
		}
		var iefp *ie.IE = nil
		fp, err := far.UpdateForwardingParameters()
		if err == nil {
			iefp = ie.NewUpdateForwardingParameters(fp...)
		}
		f.Add(NewFARUpdate(ie.NewFARID(id), ieaa, iefp))
	}
	return &f, nil, 0, 0

}

func (m *FARMapUpdate) Add(far api.FARUpdateInterface) error {
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.farmap[id]; exists {
		return fmt.Errorf("FAR %d already exists", id)
	}
	m.farmap[id] = far
	return nil
}

func (m *FARMapUpdate) IntoUpdateFAR() []*ie.IE {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r := make([]*ie.IE, len(m.farmap))

	// _ is farID, which is different from index
	i := 0
	for _, far := range m.farmap {
		r[i] = far.NewUpdateFAR()
		i++
	}
	return r
}

func (m *FARMapUpdate) Foreach(f func(api.FARUpdateInterface) error) error {
	for _, far := range m.farmap {
		err := f(far)
		if err != nil {
			return err
		}
	}
	return nil
}
