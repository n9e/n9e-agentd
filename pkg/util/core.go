// +build !windows

package util

import (
	"fmt"
	"runtime/debug"
	"syscall"

	"golang.org/x/sys/unix"
)

// SetupCoreDump enables core dumps and sets the core dump size limit based on configuration
func SetupCoreDump() error {
	debug.SetTraceback("crash")

	err := unix.Setrlimit(unix.RLIMIT_CORE, &unix.Rlimit{
		Cur: unix.RLIM_INFINITY,
		Max: unix.RLIM_INFINITY,
	})

	if err != nil {
		return fmt.Errorf("Failed to set ulimit for core dumps: %s", err)
	}

	return nil
}

func Stop() error {
	return syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
