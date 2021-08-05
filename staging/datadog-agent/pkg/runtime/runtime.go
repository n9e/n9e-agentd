package runtime

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"k8s.io/klog/v2"

	"go.uber.org/automaxprocs/maxprocs"
)

const (
	gomaxprocsKey = "GOMAXPROCS"
)

// SetMaxProcs sets the GOMAXPROCS for the go runtime to a sane value
func SetMaxProcs() {

	defer func() {
		klog.Infof("runtime: final GOMAXPROCS value is: %d", runtime.GOMAXPROCS(0))
	}()

	// This call will cause GOMAXPROCS to be set to the number of vCPUs allocated to the process
	// if the process is running in a Linux environment (including when its running in a docker / K8s setup).
	_, err := maxprocs.Set(maxprocs.Logger(klog.V(5).Infof))
	if err != nil {
		klog.Errorf("runtime: error auto-setting maxprocs: %v ", err)
	}

	if max, exists := os.LookupEnv(gomaxprocsKey); exists {
		if max == "" {
			klog.Errorf("runtime: GOMAXPROCS value was empty string")
			return
		}

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

// NumVCPU returns the number of virtualizes CPUs available to the process. It should be used instead of
// runtime.NumCPU() in virtualized environments like K8s to ensure that processes don't attempt to
// over-subscribe CPUs. For example, on a 16 vCPU machine in a docker container allocated 8 vCPUs,
// runtime.NumCPU() will return 16 but NumVCPU() will return 8.
func NumVCPU() int {
	// Value < 1 returns the current value without altering it.
	return runtime.GOMAXPROCS(0)
}