// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

type FARMapInterface interface {
	Get(key FARID) (FARInterface, error)
	Add(far FARInterface) error
	Update(far FARInterface) error
	Remove(key FARID) error
	SimulateAdd(far FARInterface) error
	SimulateUpdate(far FARInterface) error
	SimulateRemove(key FARID) error
	Foreach(func(FARInterface) error) error
}
