//go:build !windows

package executor

import (
	"os"
	"syscall"
)

func extractMaxMemory(ps *os.ProcessState) int64 {
	if ps == nil {
		return 0
	}
	if ru, ok := ps.SysUsage().(*syscall.Rusage); ok {
		return ru.Maxrss
	}

	return 0
}
