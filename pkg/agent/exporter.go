package core

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	registrymetrics "github.com/n9e/n9e-agentd/pkg/registry/metrics"
	"github.com/n9e/n9e-agentd/pkg/telemetry"
	"k8s.io/klog/v2"
)

const (
	shutDownTimeout        = 5 * time.Second
	defaultKeepAlivePeriod = 3 * time.Minute
)

type server struct {
	address string
	handler http.Handler
}

func (p *module) startExporter() error {
	cf := p.config.Exporter
	if cf.Port <= 0 {
		return nil
	}

	mux := http.NewServeMux()

	if cf.Docs {
		mux.Handle("/docs/metrics", registrymetrics.Handler())
	}

	if cf.Metrics {
		mux.Handle("/metrics", telemetry.Handler())
	}

	if cf.Expvar {
		mux.Handle("/vars", expvar.Handler())
	}

	if cf.Pprof {
		mux.HandleFunc("/debug/pprof", redirectTo("/debug/pprof/"))
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	server := &server{
		address: fmt.Sprintf("127.0.0.1:%d", cf.Port),
		handler: mux,
	}

	return server.start(p.ctx)
}

func (s *server) start(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    listener.Addr().String(),
		Handler: s.handler,
	}

	go func() {
		<-ctx.Done()

		ctx2, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
		server.Shutdown(ctx2)
		cancel()
	}()

	go func() {
		err := server.Serve(tcpKeepAliveListener{listener})

		msg := fmt.Sprintf("Stopped listening on %s", listener.Addr().String())
		select {
		case <-ctx.Done():
			klog.Info(msg)
		default:
			panic(fmt.Sprintf("%s due to error: %v", msg, err))
		}
	}()

	return nil
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// Copied from Go 1.7.2 net/http/server.go
type tcpKeepAliveListener struct {
	net.Listener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	c, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}
	if tc, ok := c.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(defaultKeepAlivePeriod)
	}
	return c, nil
}

// redirectTo redirects request to a certain destination.
func redirectTo(to string) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		http.Redirect(rw, req, to, http.StatusFound)
	}
}
