// Copyright Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package api

type SessionsMapInterface interface {
	Add(session PFCPSessionInterface) error
	GetPFCPSessions() []PFCPSessionInterface
	GetPFCPSession(localIP string, seid SEID) (PFCPSessionInterface, error)
}
