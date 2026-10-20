//go:build windows

package teacher

import "os/exec"

func exitCode(err *exec.ExitError) int {
	return err.ExitCode()
}
