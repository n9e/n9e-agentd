package options

import (
	"context"

	"github.com/yubo/golib/proc"
	"k8s.io/klog/v2"
)

func InstallReporter() {
	proc.RegisterHooks(reporterHookOps)
}

const (
	moduleName = "build.reporter"
)

var (
	_reporter       = &reporter{}
	reporterHookOps = []proc.HookOps{{
		Hook:     _reporter.start,
		Owner:    moduleName,
		HookNum:  proc.ACTION_START,
		Priority: proc.PRI_MODULE,
	}, {
		Hook:     _reporter.stop,
		Owner:    moduleName,
		HookNum:  proc.ACTION_STOP,
		Priority: proc.PRI_MODULE,
	}}
)

type reporter struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (p *reporter) start(ctx context.Context) error {
	if p.cancel != nil {
		p.cancel()
	}
	p.ctx, p.cancel = context.WithCancel(ctx)

	reporter := &buildReporter{}

	if err := reporter.Start(); err != nil {
		return err
	}

	go func() {
		<-p.ctx.Done()
		reporter.Stop()
	}()

	return nil
}

func (p *reporter) stop(ctx context.Context) error {
	klog.Info("stop")
	p.cancel()
	return nil
}
