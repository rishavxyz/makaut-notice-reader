package utils

import (
	"net"
	"time"
)

func IsOffline(timeout ...time.Duration) bool {
	var tout time.Duration

	if len(timeout) > 0 {
		tout = timeout[0]
	} else {
		tout = 8e9 // 8s
	}
	var _, err = net.DialTimeout("tcp", "golang.org:80", tout)

	if err != nil {
		return true
	}
	return false
}
