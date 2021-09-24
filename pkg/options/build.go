// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package options

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/DataDog/datadog-agent/pkg/telemetry"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/util"
	"k8s.io/klog/v2"
)

var (
	// Revision is the VCS revision associated with this build. Overridden using ldflags
	// at compile time. Example:
	// $ go build -ldflags "-X github.com/m3db/m3/src/x/instrument.Revision=abcdef" ...
	// Adapted from: https://www.atatus.com/blog/golang-auto-build-versioning/
	Revision = "unknown"

	// Branch is the VCS branch associated with this build.
	Branch = "unknown"

	// Version is the version associated with this build.
	Version = "unknown"

	// Builder is the username this build was created
	Builder = "unknown"

	// BuildDate is the date this build was created.
	BuildDate = "unknown"

	// BuildTimeUnix is the seconds since epoch representing the date this build was created.
	BuildTimeUnix = "0"

	// LogBuildInfoAtStartup controls whether we log build information at startup. If its
	// set to a non-empty string, we log the build information at process startup.
	LogBuildInfoAtStartup string

	// GoVersion is the current runtime version.
	GoVersion = runtime.Version()

	goOs   = runtime.GOOS
	goArch = runtime.GOARCH
)

var (
	errAlreadySetted     = errors.New("options already setted")
	errNotSetted         = errors.New("options not setted")
	errAlreadyStarted    = errors.New("reporter already started")
	errNotStarted        = errors.New("reporter not started")
	errBuildTimeNegative = errors.New("reporter build time must be non-negative")
)

// LogBuildInfo logs the build information to the provided logger.
func LogBuildInfo() {
	klog.Infof("Go Runtime version: %s\n", GoVersion)
	klog.Infof("OS:                 %s\n", goOs)
	klog.Infof("Arch:               %s\n", goArch)
	klog.Infof("Build Version:      %s\n", Version)
	klog.Infof("Build Revision:     %s\n", Revision)
	klog.Infof("Build Branch:       %s\n", Branch)
	klog.Infof("Build User:         %s\n", Builder)
	klog.Infof("Build Date:         %s\n", BuildDate)
	klog.Infof("Build TimeUnix:     %s\n", BuildTimeUnix)
}

type buildReporter struct {
	sync.Mutex

	buildInfoGauge telemetry.Gauge
	buildAgeGauge  telemetry.Gauge

	buildTime time.Time
	active    bool
	closeCh   chan struct{}
	doneCh    chan struct{}
}

func (b *buildReporter) Start() error {
	const (
		base    = 10
		bitSize = 64
	)
	sec, err := strconv.ParseInt(BuildTimeUnix, base, bitSize)
	if err != nil {
		return err
	}
	if sec < 0 {
		return errBuildTimeNegative
	}
	buildTime := time.Unix(sec, 0)

	b.Lock()
	defer b.Unlock()
	if b.active {
		return errAlreadyStarted
	}
	b.buildTime = buildTime
	b.active = true
	b.closeCh = make(chan struct{})
	b.doneCh = make(chan struct{})
	go b.report()
	return nil
}

func (b *buildReporter) _report() {
	b.buildInfoGauge.Set(1.0, Revision, Branch, BuildDate, Version, GoVersion)
	b.buildAgeGauge.Set(float64(time.Since(b.buildTime)), Revision, Branch, BuildDate, Version, GoVersion)
}

func (b *buildReporter) report() {
	tags := []string{"revision", "branch", "build_date", "build_version", "go_version"}
	b.buildInfoGauge = telemetry.NewGauge("agentd", "build_information", tags, "agentd build information")
	b.buildAgeGauge = telemetry.NewGauge("agentd", "build_age", tags, "agentd build information")

	b._report()

	ticker := time.NewTicker(time.Second * 10)
	defer func() {
		close(b.doneCh)
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			b._report()
		case <-b.closeCh:
			return
		}
	}
}

func (b *buildReporter) Stop() error {
	b.Lock()
	defer b.Unlock()
	if !b.active {
		return errNotStarted
	}
	close(b.closeCh)
	<-b.doneCh
	b.active = false
	return nil
}

type ProcVersion struct {
	Version   string `json:"version,omitempty"`
	Release   string `json:"release,omitempty"`
	Git       string `json:"git,omitempty"`
	Go        string `json:"go,omitempty"`
	Os        string `json:"os,omitempty"`
	Arch      string `json:"arch,omitempty"`
	Builder   string `json:"builder,omitempty"`
	BuildTime int64  `json:"buildTime,omitempty" out:",date"`
}

func (p ProcVersion) String() string {
	return util.Prettify(p)
}

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version, git commit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return VersionCmd(cmd, args)
		},
	}
	return cmd
}

func VersionCmd(cmd *cobra.Command, args []string) error {
	fmt.Printf("Go Runtime version: %s\n", GoVersion)
	fmt.Printf("OS:                 %s\n", goOs)
	fmt.Printf("Arch:               %s\n", goArch)
	fmt.Printf("Build Version:      %s\n", Version)
	fmt.Printf("Build Revision:     %s\n", Revision)
	fmt.Printf("Build Branch:       %s\n", Branch)
	fmt.Printf("Build User:         %s\n", Builder)
	fmt.Printf("Build Date:         %s\n", BuildDate)
	fmt.Printf("Build TimeUnix:     %s\n", BuildTimeUnix)
	return nil
}
