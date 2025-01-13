// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

import "github.com/wmnsk/go-pfcp/ie"

type FARMapUpdateInterface interface {
	Add(far FARUpdateInterface) error
	IntoUpdateFAR() []*ie.IE
	Foreach(f func(FARUpdateInterface) error) error
}
