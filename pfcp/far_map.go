// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/nextmn/go-pfcp-networking/internal/loglevel"
	"github.com/nextmn/go-pfcp-networking/pfcp/api"

	"github.com/wmnsk/go-pfcp/ie"
)

type farmapInternal = map[api.FARID]api.FARInterface

type FARMap struct {
	farmap farmapInternal
	mu     sync.RWMutex
}

func (m *FARMap) Foreach(f func(api.FARInterface) error) error {
	for _, far := range m.farmap {
		err := f(far)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *FARMap) Get(key api.FARID) (api.FARInterface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if far, exists := m.farmap[key]; exists {
		return far, nil
	}
	return nil, fmt.Errorf("FAR %d does not exist", key)
}

func (m *FARMap) Add(far api.FARInterface) error {
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

func (m *FARMap) SimulateAdd(far api.FARInterface) error {
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.farmap[id]; exists {
		return fmt.Errorf("FAR %d already exists", id)
	}
	return nil
}

func (m *FARMap) Update(farUpdate api.FARUpdateInterface) error {
	// FIXME: add Context
	slog.Log(context.TODO(), loglevel.Trace, "Inside farmap.Update()")
	// only present fields are replaced
	id, err := farUpdate.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if far, exists := m.farmap[id]; !exists {
		slog.Log(context.TODO(), loglevel.Trace, "Updating FAR: this FAR id does not exist",
			"far-id", id,
			"current-map", m.farmap,
		)
		return fmt.Errorf("FAR %d does not exist", id)
	} else {
		slog.Log(context.TODO(), loglevel.Trace, "Updating FAR", "far-id", id)
		return far.Update(farUpdate)
	}
}

func (m *FARMap) SimulateUpdate(far api.FARUpdateInterface) error {
	// FIXME: add Context
	slog.Log(context.TODO(), loglevel.Trace, "Inside farmap.SimulateUpdate()")
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.farmap[id]; !exists {
		slog.Log(context.TODO(), loglevel.Trace, "Simulate updating FAR: this FAR id does not exist",
			"far-id", id,
			"current-map", m.farmap,
		)
		return fmt.Errorf("FAR %d does not exist", id)
	}
	slog.Log(context.TODO(), loglevel.Trace, "Simulate updating FAR",
		"far-id", id,
		"current-map", m.farmap,
	)
	return nil
}
func (m *FARMap) Remove(key api.FARID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.farmap[key]; !exists {
		return fmt.Errorf("FAR %d does not exist", key)
	} else {
		delete(m.farmap, key)
		return nil
	}
}
func (m *FARMap) SimulateRemove(key api.FARID) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.farmap[key]; !exists {
		return fmt.Errorf("FAR %d does not exist", key)
	}
	return nil
}
func (m *FARMap) NewCreateFARs() []*ie.IE {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f := make([]*ie.IE, 0)
	for _, far := range m.farmap {
		f = append(f, far.NewCreateFAR())
	}
	return f
}

func NewFARMap(fars []*ie.IE) (farmap *FARMap, cause uint8, offendingIE uint16, err error) {
	f := FARMap{
		farmap: make(farmapInternal),
		mu:     sync.RWMutex{},
	}
	for _, far := range fars {
		id, err := far.FARID()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, ie.CauseInvalidLength, ie.FARID, err
			case ie.ErrIENotFound:
				return nil, ie.CauseMandatoryIEMissing, ie.FARID, err
			default:
				return nil, ie.CauseMandatoryIEIncorrect, ie.CreateFAR, err
			}
		}
		aa, err := far.ApplyAction()
		if err != nil {
			switch err {
			case io.ErrUnexpectedEOF:
				return nil, ie.CauseInvalidLength, ie.ApplyAction, err
			case ie.ErrIENotFound:
				return nil, ie.CauseMandatoryIEMissing, ie.ApplyAction, err
			default:
				return nil, ie.CauseMandatoryIEIncorrect, ie.CreateFAR, err
			}
		}

		// This IE shall be present when the Apply Action requests
		// the packets to be forwarded. It may be present otherwise.
		mustHaveFP := false
		hasFP := false
		if far.HasFORW() {
			mustHaveFP = true
		}
		fp, err := far.ForwardingParameters()
		if err == nil {
			hasFP = true
		}
		if mustHaveFP && !hasFP {
			return nil, ie.CauseMandatoryIEIncorrect, ie.CreateFAR, err
		}

		if !hasFP {
			err = f.Add(NewFAR(ie.NewFARID(id), ie.NewApplyAction(aa...), nil))
		} else {
			err = f.Add(NewFAR(ie.NewFARID(id), ie.NewApplyAction(aa...), ie.NewForwardingParameters(fp...)))
		}
		if err != nil {
			return nil, ie.CauseMandatoryIEIncorrect, ie.CreateFAR, err
		}
	}
	return &f, 0, 0, nil

}

func (m *FARMap) IntoCreateFAR() []*ie.IE {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r := make([]*ie.IE, len(m.farmap))

	// _ is farID, which is different from index
	i := 0
	for _, far := range m.farmap {
		r[i] = far.NewCreateFAR()
		i++
	}
	return r
}
