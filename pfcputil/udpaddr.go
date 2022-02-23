// Copyright 2022 Louis Royer and the go-pfcp-networking contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package pfcputil

import (
	"fmt"
	"strings"
)

func CreateUDPAddr(ipaddr string, port string) string {
	if strings.Count(ipaddr, ":") > 0 {
		return fmt.Sprintf("[%s]:%s", ipaddr, port)
	}
	return fmt.Sprintf("%s:%s", ipaddr, port)
}
