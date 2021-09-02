// +build !windows

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func DefaultConfigfile() string {
	root, err := ResolveRootPath("")
	if err != nil {
		return ""
	}

	return filepath.Join(root, "etc", "agentd.yaml")
}

func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// writes auth token(s) to a file that is only readable/writable by the user running the agent
func saveAuthToken(token, tokenPath string) error {
	return ioutil.WriteFile(tokenPath, []byte(token), 0600)
}
