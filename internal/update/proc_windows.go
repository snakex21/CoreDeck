//go:build windows

package update

import (
	"os/exec"
	"syscall"
)

const (
	createNewProcessGroup  = 0x00000200
	createBreakawayFromJob = 0x01000000
)

// configureDetached sets creation flags; it reports whether breakaway was
// requested (a retry without it is then possible).
func configureDetached(cmd *exec.Cmd, breakaway bool) bool {
	flags := uint32(createNewProcessGroup)
	if breakaway {
		flags |= createBreakawayFromJob
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: flags}
	return breakaway
}
