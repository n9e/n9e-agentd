package core

import (
	"runtime"
	"strconv"
	"strings"

	"go.uber.org/automaxprocs/maxprocs"
	"k8s.io/klog/v2"
)

// setMaxProcs sets the GOMAXPROCS for the go runtime to a sane value
func setMaxProcs(max string) {

	defer func() {
		klog.Infof("runtime: final GOMAXPROCS value is: %d", runtime.GOMAXPROCS(0))
	}()

	// This call will cause GOMAXPROCS to be set to the number of vCPUs allocated to the process
	// if the process is running in a Linux environment (including when its running in a docker / K8s setup).
	_, err := maxprocs.Set(maxprocs.Logger(klog.V(5).Infof))
	if err != nil {
		klog.Errorf("runtime: error auto-setting maxprocs: %v ", err)
	}

	if len(max) > 0 {
		_, err = strconv.Atoi(max)
		if err == nil {
			// Go runtime will already have parsed the integer and set it properly.
			return
		}

		if strings.HasSuffix(max, "m") {
			// Value represented as millicpus.
			trimmed := strings.TrimSuffix(max, "m")
			milliCPUs, err := strconv.Atoi(trimmed)
			if err != nil {
				klog.Errorf("runtime: error parsing GOMAXPROCS milliCPUs value: %v", max)
				return
			}

			cpus := milliCPUs / 1000
			if cpus > 0 {
				klog.Infof("runtime: honoring GOMAXPROCS millicpu configuration: %v, setting GOMAXPROCS to: %d", max, cpus)
				runtime.GOMAXPROCS(cpus)
			} else {
				klog.Infof(
					"runtime: GOMAXPROCS millicpu configuration: %s was less than 1, setting GOMAXPROCS to 1",
					max)
				runtime.GOMAXPROCS(1)
			}
			return
		}

		klog.Errorf(
			"runtime: unhandled GOMAXPROCS value: %s", max)
	}
}
