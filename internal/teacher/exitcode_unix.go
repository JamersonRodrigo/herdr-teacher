//go:build !windows

package teacher

import (
	"os/exec"
	"syscall"
)

func exitCode(err *exec.ExitError) int {
	if ws, ok := err.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return err.ExitCode()
}
