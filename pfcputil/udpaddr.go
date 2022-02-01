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
