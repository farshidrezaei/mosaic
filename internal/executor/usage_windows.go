//go:build windows

package executor

import "os"

func extractMaxMemory(_ *os.ProcessState) int64 {
	return 0
}
