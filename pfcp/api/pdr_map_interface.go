// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import "github.com/wmnsk/go-pfcp/ie"

type PDRMapInterface interface {
	Get(key PDRID) (PDRInterface, error)
	Add(pdr PDRInterface) error
	Update(pdr PDRInterface) error
	Remove(key PDRID) error
	SimulateAdd(pdr PDRInterface) error
	SimulateUpdate(pdr PDRInterface) error
	SimulateRemove(key PDRID) error
	GetSortedPDRIDs() []PDRID
	Foreach(func(PDRInterface) error) error
	IntoCreatePDR() []*ie.IE
	IntoUpdatePDR() []*ie.IE
}
