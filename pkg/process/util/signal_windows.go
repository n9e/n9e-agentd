// +build windows

package util

import (
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
)

// HandleSignals tells us whether we should exit.
func HandleSignals(exit chan struct{}) {
	sigIn := make(chan os.Signal, 100)
	signal.Notify(sigIn, syscall.SIGINT, syscall.SIGTERM)
	// unix only in all likelihood;  but we don't care.
	for sig := range sigIn {
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			klog.Infof("Caught signal '%s'; terminating.", sig)
			close(exit)
			return
		}
	}
}
