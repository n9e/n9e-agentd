//go:build !windows
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
	"k8s.io/klog/v2"
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

func ValidateDirs(paths ...string) error {
	for _, path := range paths {
		if err := ValidateDir(path); err != nil {
			return err
		}
	}

	return nil
}

func ValidateDir(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
		klog.Infof("directory %s does not exist, created", path)
		return nil
	}
	if fi.IsDir() {
		return nil
	}

	return fmt.Errorf("Cannot create directory since %s already exists", path)
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
