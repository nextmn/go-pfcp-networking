// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcp_networking

import (
	"fmt"
	"sync"

	"github.com/louisroyer/go-pfcp-networking/pfcp/api"
)

type associationsMap = map[string]api.PFCPAssociationInterface
type AssociationsMap struct {
	associations   associationsMap
	muAssociations sync.RWMutex
}

func NewAssociationsMap() AssociationsMap {
	return AssociationsMap{
		associations:   make(associationsMap),
		muAssociations: sync.RWMutex{},
	}
}

// Remove an association from the association table
func (a *AssociationsMap) Remove(association api.PFCPAssociationInterface) error {
	nid, err := association.NodeID().NodeID()
	if err != nil {
		return err
	}
	a.muAssociations.Lock()
	defer a.muAssociations.Unlock()
	delete(a.associations, nid)
	return nil
}

// Add an association to the association table
func (a *AssociationsMap) Add(association api.PFCPAssociationInterface) error {
	nid, err := association.NodeID().NodeID()
	if err != nil {
		return err
	}
	if _, exists := a.associations[nid]; exists {
		// Only one association shall be setup between given pair of CP and UP functions.
		return fmt.Errorf("Association already exist.")
	}
	a.muAssociations.Lock()
	defer a.muAssociations.Unlock()
	a.associations[nid] = association
	return nil
}

// Returns true if the association does not exist
func (a *AssociationsMap) CheckNonExist(nid string) bool {
	a.muAssociations.RLock()
	defer a.muAssociations.RUnlock()
	if _, exists := a.associations[nid]; exists {
		return false
	}
	return true
}

// Returns a copy of an existing PFCP Association
func (a *AssociationsMap) Get(nid string) (association api.PFCPAssociationInterface, err error) {
	a.muAssociations.RLock()
	defer a.muAssociations.RUnlock()
	if asso, exists := a.associations[nid]; exists {
		return asso, nil
	}
	return nil, fmt.Errorf("Association does not exist.")
}

// Update a Association
func (a *AssociationsMap) Update(association api.PFCPAssociationInterface) error {
	nid, err := association.NodeID().NodeID()
	if err != nil {
		return err
	}
	a.muAssociations.Lock()
	defer a.muAssociations.Unlock()
	a.associations[nid] = association
	return nil
}
