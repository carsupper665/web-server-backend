// common/portScanner.go
package common

import (
	"fmt"
	"net"
)

// CheckPortAvailable 試著綁定 TCP port，確認沒人用
func CheckPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// 找一個指定範圍內第一個 free port（你可以改成靜態分配邏輯）
func PickStaticPort(start, end int) (int, error) {
	for p := start; p <= end; p++ {
		if CheckPortAvailable(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", start, end)
}
