// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import "github.com/wmnsk/go-pfcp/ie"

type FARUpdateInterface interface {
	ID() (FARID, error)
	ApplyAction() *ie.IE
	UpdateForwardingParameters() *ie.IE
	NewUpdateFAR() *ie.IE
}
