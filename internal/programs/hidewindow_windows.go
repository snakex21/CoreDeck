//go:build windows

package programs

import (
	"os/exec"
	"syscall"
)

const createNoWindow = 0x08000000

// hidden configures an *exec.Cmd so the spawned console process does not pop a
// visible console window.
func hidden(cmd *exec.Cmd) *exec.Cmd {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	return cmd
}
