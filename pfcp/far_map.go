// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"io"
	"sync"

	"github.com/nextmn/go-pfcp-networking/pfcp/api"
	"github.com/sirupsen/logrus"
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
	return nil, fmt.Errorf("FAR %d does not exist.", key)
}

func (m *FARMap) Add(far api.FARInterface) error {
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.farmap[id]; exists {
		return fmt.Errorf("FAR %d already exists.", id)
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
		return fmt.Errorf("FAR %d already exists.", id)
	}
	return nil
}

func (m *FARMap) Update(far api.FARInterface) error {
	logrus.Trace("Inside farmap.Update()")
	// only present fields are replaced
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.farmap[id]; !exists {
		logrus.WithFields(logrus.Fields{"far-id": id, "current_map": m.farmap}).Trace("Updating FAR: this FAR id does not exist")
		return fmt.Errorf("FAR %d does not exist.", id)
	} else {
		logrus.WithFields(logrus.Fields{"far-id": id}).Trace("Updating FAR")
		if far.ApplyAction() != nil {
			m.farmap[id].SetApplyAction(far.ApplyAction())
			logrus.WithFields(logrus.Fields{"far-id": id}).Trace("Updating FAR Apply Action")
		}
		// XXX: update fields in forwarding parameters instead of replacing
		if fp, err := far.ForwardingParameters(); err == nil {
			if fp == nil {
				logrus.Warn("Removing forwarding parameters. aborting")
				return nil
			}
			m.farmap[id].SetForwardingParameters(fp)
			logrus.WithFields(logrus.Fields{"far-id": id}).Trace("Updating FAR Forwarding Parameters")
		} else {
			logrus.WithFields(logrus.Fields{"far-id": id}).Trace("Updating FAR but not Forwarding Parameters")
		}

		return nil
	}
}

func (m *FARMap) SimulateUpdate(far api.FARInterface) error {
	logrus.Trace("Inside farmap.SimulateUpdate()")
	id, err := far.ID()
	if err != nil {
		return err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.farmap[id]; !exists {
		logrus.WithFields(logrus.Fields{"far-id": id, "current_map": m.farmap}).Trace("Simulate Updating FAR: this FAR id does not exist")
		return fmt.Errorf("FAR %d does not exist.", id)
	}
	logrus.WithFields(logrus.Fields{"far-id": id, "current_map": m.farmap}).Trace("Simulate Updating FAR: exist")
	return nil
}
func (m *FARMap) Remove(key api.FARID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.farmap[key]; !exists {
		return fmt.Errorf("FAR %d does not exist.", key)
	} else {
		delete(m.farmap, key)
		return nil
	}
}
func (m *FARMap) SimulateRemove(key api.FARID) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.farmap[key]; !exists {
		return fmt.Errorf("FAR %d does not exist.", key)
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

func NewFARMap(fars []*ie.IE) (farmap *FARMap, err error, cause uint8, offendingIE uint16) {
	f := FARMap{
		farmap: make(farmapInternal),
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
			return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreateFAR
		}

		if !hasFP {
			err = f.Add(NewFAR(ie.NewFARID(id), ie.NewApplyAction(aa...), nil))
		} else {
			err = f.Add(NewFAR(ie.NewFARID(id), ie.NewApplyAction(aa...), ie.NewForwardingParameters(fp...)))
		}
		if err != nil {
			return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreateFAR
		}
	}
	return &f, nil, 0, 0

}

func NewFARMapUpdate(fars []*ie.IE) (*FARMap, error, uint8, uint16) {
	f := FARMap{
		farmap: make(farmapInternal),
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
			iefp = ie.NewForwardingParameters(fp...)
		}
		f.Add(NewFAR(ie.NewFARID(id), ieaa, iefp))
	}
	return &f, nil, 0, 0

}
