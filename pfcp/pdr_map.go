// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
	"github.com/wmnsk/go-pfcp/ie"
)

type pdrmapInternal = map[api.PDRID]api.PDRInterface

type PDRMap struct {
	mu     sync.RWMutex
	pdrmap pdrmapInternal

	muArray   sync.Mutex
	isSorted  bool
	isUpdated bool
	sortedIDs []api.PDRID // sortedID is an array with PDR IDs in search order
}

func (m *PDRMap) Foreach(f func(api.PDRInterface) error) error {
	for _, pdr := range m.pdrmap {
		err := f(pdr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Begin - Functions used internally to sort PDR IDs
// calling them is unsafe
func (m *PDRMap) Less(i, j int) bool {
	// element with highest precedence (lowest value in Precedence IE) should be sorted first
	pdridi := m.sortedIDs[i]
	pi, err := m.pdrmap[pdridi].Precedence()
	if err != nil {
		return false
	}
	pdridj := m.sortedIDs[j]
	pj, err := m.pdrmap[pdridj].Precedence()
	if err != nil {
		return true
	}
	return pi < pj
}

func (m *PDRMap) Len() int {
	return len(m.sortedIDs)
}
func (m *PDRMap) Swap(i, j int) {
	m.sortedIDs[i], m.sortedIDs[j] = m.sortedIDs[j], m.sortedIDs[i]
}

// End - Functions used internally to sort PDR IDs

func (m *PDRMap) updateArray() {
	m.muArray.Lock()
	defer m.muArray.Unlock()
	if m.isUpdated {
		// the array has been updated before we got the lock
		return
	}
	keys := make([]api.PDRID, 0)
	for k := range m.pdrmap {
		keys = append(keys, k)
	}
	m.isUpdated = true
	m.isSorted = false
	m.sortedIDs = keys
}

func (m *PDRMap) sortArray() {
	m.muArray.Lock()
	defer m.muArray.Unlock()
	if m.isSorted {
		// the array has been sorted before we got the lock
		return
	}
	sort.Sort(m)
	m.isSorted = true
}

func (m *PDRMap) GetSortedPDRIDs() []api.PDRID {
	m.mu.RLock()
	defer m.mu.RUnlock()
	switch {
	case !m.isUpdated:
		// Recompute keys
		m.updateArray()
		fallthrough // now perform sort
	case !m.isSorted:
		// Sort
		m.sortArray()
		fallthrough // now return
	default:
		return m.sortedIDs
	}
}

func (m *PDRMap) NewCreatePDRs() []*ie.IE {
	p := make([]*ie.IE, 0)
	for _, pdr := range m.pdrmap {
		p = append(p, pdr.NewCreatePDR())
	}
	return p
}
func (m *PDRMap) Get(key api.PDRID) (api.PDRInterface, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if pdr, exists := m.pdrmap[key]; exists {
		return pdr, nil
	}
	return nil, fmt.Errorf("PDR %d does not exist.", key)

}
func (m *PDRMap) Add(pdr api.PDRInterface) error {
	id, err := pdr.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pdrmap[id]; exists {
		return fmt.Errorf("PDR %d already exists.", id)
	}
	if m.isUpdated {
		// keep the array updated (but not sorted)
		m.muArray.Lock()
		defer m.muArray.Unlock()
		m.sortedIDs = append(m.sortedIDs, id)
		m.isSorted = false
	}
	m.pdrmap[id] = pdr
	return nil
}

func (m *PDRMap) SimulateAdd(pdr api.PDRInterface) error {
	id, err := pdr.ID()
	if err != nil {
		return err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.pdrmap[id]; exists {
		return fmt.Errorf("PDR %d already exists.", id)
	}
	return nil
}

func (m *PDRMap) Update(pdr api.PDRInterface) error {
	id, err := pdr.ID()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pdrmap[id]; !exists {
		return fmt.Errorf("PDR %d does not exist.", id)
	} else {
		delete(m.pdrmap, id)
		m.muArray.Lock()
		defer m.muArray.Unlock()
		m.isSorted = false
		m.isUpdated = false
		m.pdrmap[id] = pdr
		return nil
	}
}

func (m *PDRMap) SimulateUpdate(pdr api.PDRInterface) error {
	id, err := pdr.ID()
	if err != nil {
		return err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.pdrmap[id]; !exists {
		return fmt.Errorf("PDR %d does not exist.", id)
	}
	return nil
}

func (m *PDRMap) Remove(key api.PDRID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.pdrmap[key]; !exists {
		return fmt.Errorf("PDR %d does not exist.", key)
	} else {
		delete(m.pdrmap, key)
		m.muArray.Lock()
		defer m.muArray.Unlock()
		m.isUpdated = false
		return nil
	}
}

func (m *PDRMap) SimulateRemove(key api.PDRID) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, exists := m.pdrmap[key]; !exists {
		return fmt.Errorf("PDR %d does not exist.", key)
	}
	return nil
}

func NewPDRMap(pdrs []*ie.IE) (pdr *PDRMap, err error, cause uint8, offendingIE uint16) {
	p := PDRMap{
		pdrmap:    make(pdrmapInternal),
		mu:        sync.RWMutex{},
		muArray:   sync.Mutex{},
		isSorted:  false,
		isUpdated: false,
		sortedIDs: make([]api.PDRID, 0),
	}
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

		err = p.Add(NewPDR(
			ie.NewPDRID(id),
			ie.NewPDI(pdi...),
			ie.NewPrecedence(precedence),
			ie.NewFARID(farid),
			ohrIE,
		))
		if err != nil {
			return nil, err, ie.CauseMandatoryIEIncorrect, ie.CreatePDR
		}
	}
	return &p, nil, 0, 0
}
