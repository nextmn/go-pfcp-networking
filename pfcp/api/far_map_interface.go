// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import (
	"github.com/wmnsk/go-pfcp/ie"
)

type FARMapInterface interface {
	Get(key FARID) (FARInterface, error)
	Add(far FARInterface) error
	Update(far FARUpdateInterface) error
	Remove(key FARID) error
	SimulateAdd(far FARInterface) error
	SimulateUpdate(far FARUpdateInterface) error
	SimulateRemove(key FARID) error
	Foreach(func(FARInterface) error) error
	IntoCreateFAR() []*ie.IE
}
