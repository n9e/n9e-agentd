package apiserver

import (
	"fmt"

	"github.com/DataDog/datadog-agent/pkg/logs"
	"github.com/DataDog/datadog-agent/pkg/logs/diagnostic"
	"github.com/yubo/apiserver/pkg/watch"
	"k8s.io/klog/v2"
)

type logWatcher struct {
	logChan  <-chan string
	done     chan struct{}
	result   chan watch.Event
	stopped  bool
	running  bool
	receiver *diagnostic.BufferedMessageReceiver
}

func newLogsWatch(filters *diagnostic.Filters) (*logWatcher, error) {
	receiver := logs.GetMessageReceiver()
	if receiver == nil {
		klog.Info("Logs agent is not running - can't stream logs")
		return nil, fmt.Errorf("The logs agent is not running")
	}

	if !receiver.SetEnabled(true) {
		klog.Info("Logs are already streaming. Dropping connection.")
		return nil, fmt.Errorf("Another client is already streaming logs.")
	}
	done := make(chan struct{})

	logChan := receiver.Filter(filters, done)

	return &logWatcher{
		logChan:  logChan,
		done:     done,
		result:   make(chan watch.Event, 100),
		receiver: receiver,
	}, nil
}

func (p *logWatcher) Start() error {
	if p.running {
		return fmt.Errorf("logWatcher is already running")
	}
	p.running = true
	go func() {
		for {
			select {
			case log := <-p.logChan:
				p.result <- watch.Event{Type: watch.Added, Object: &log}
			case <-p.done:
				return
			}
		}
	}()

	return nil
}

func (p *logWatcher) Stop() {
	if !p.stopped {
		klog.V(4).Infof("Stopping log watcher.")
		close(p.result)
		close(p.done)
		p.receiver.SetEnabled(false)
		p.stopped = true
		p.running = false
	}
}

func (p *logWatcher) ResultChan() <-chan watch.Event {
	return p.result
}
