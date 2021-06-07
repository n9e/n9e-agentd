package apiserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	gorilla "github.com/gorilla/mux"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/n9e/n9e-agentd/pkg/apiserver/check"
	"github.com/n9e/n9e-agentd/pkg/apiserver/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func (p *module) grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	deadline := time.Now().Add(p.config.ServerTimeout)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			conn := GetConnection(r)
			_ = conn.SetWriteDeadline(deadline)

			otherHandler.ServeHTTP(w, r)
		}
	})
}

func (p *module) startServer() (err error) {
	p.initializeTLS()

	// gRPC server
	mux := http.NewServeMux()
	opts := []grpc.ServerOption{
		grpc.Creds(credentials.NewClientTLSFromCert(p.tlsCertPool, p.tlsAddr)),
		grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(p.grpcAuth)),
		grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(p.grpcAuth)),
	}

	s := grpc.NewServer(opts...)
	pb.RegisterAgentServer(s, &server{hostname: p.hostname})
	pb.RegisterAgentSecureServer(s, &serverSecure{})

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: p.tlsAddr,
		RootCAs:    p.tlsCertPool,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// starting grpc gateway
	gwmux := runtime.NewServeMux()
	err = pb.RegisterAgentHandlerFromEndpoint(
		p.ctx, gwmux, p.tlsAddr, dopts)
	if err != nil {
		panic(err)
	}

	err = pb.RegisterAgentSecureHandlerFromEndpoint(
		p.ctx, gwmux, p.tlsAddr, dopts)
	if err != nil {
		panic(err)
	}

	// Setup multiplexer
	// create the REST HTTP router
	agentMux := gorilla.NewRouter()
	checkMux := gorilla.NewRouter()
	// Validate token for every request
	agentMux.Use(validateToken)
	checkMux.Use(validateToken)

	mux.Handle("/agent/", http.StripPrefix("/agent", p.SetupHandlers(agentMux)))
	mux.Handle("/check/", http.StripPrefix("/check", check.SetupHandlers(checkMux)))
	mux.Handle("/", gwmux)

	srv := &http.Server{
		Addr:    p.tlsAddr,
		Handler: p.grpcHandlerFunc(s, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*p.tlsKeyPair},
			NextProtos:   []string{"h2"},
		},
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			// Store the connection in the context so requests can reference it if needed
			return context.WithValue(ctx, contextKeyConn, c)
		},
	}

	tlsListener := tls.NewListener(p.listener, srv.TLSConfig)

	go srv.Serve(tlsListener) //nolint:errcheck
	return nil

}

// getIPCAddressPort returns a listening connection
func (p config) getIPCAddressPort() (string, error) {
	address, err := p.getIPCAddress()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v:%v", address, p.CmdPort), nil
}

// GetIPCAddress returns the IPC address or an error if the address is not local
func (p config) getIPCAddress() (string, error) {
	address := p.IpcAddress
	if address == "localhost" {
		return address, nil
	}
	ip := net.ParseIP(address)
	if ip == nil {
		return "", fmt.Errorf("apiserver.ipcAddress was set to an invalid IP address: %s", address)
	}
	for _, cidr := range []string{
		"127.0.0.0/8", // IPv4 loopback
		"::1/128",     // IPv6 loopback
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			return "", err
		}
		if block.Contains(ip) {
			return address, nil
		}
	}
	return "", fmt.Errorf("ipc_address was set to a non-loopback IP address: %s", address)
}
