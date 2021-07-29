package grpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var defaultBackoffConfig = backoff.Config{
	BaseDelay:  1.0 * time.Second,
	Multiplier: 1.1,
	Jitter:     0.2,
	MaxDelay:   2 * time.Second,
}

// defaultAgentDialOpts default dial options to the main agent which blocks and retries based on the backoffConfig
var defaultAgentDialOpts = []grpc.DialOption{
	grpc.WithConnectParams(grpc.ConnectParams{Backoff: defaultBackoffConfig}),
	grpc.WithBlock(),
}

// GetDDAgentClient creates a pb.AgentClient for IPC with the main agent via gRPC. This call is blocking by default, so
// it is up to the caller to supply a context with appropriate timeout/cancel options
// func GetDDAgentClient(cf *config.Config, ctx context.Context, opts ...grpc.DialOption) (pb.AgentClient, error) {
// 	// This is needed as the server hangs when using "grpc.WithInsecure()"
// 	tlsConf := tls.Config{InsecureSkipVerify: true}
//
// 	if len(opts) == 0 {
// 		opts = defaultAgentDialOpts
// 	}
//
// 	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConf)))
//
// 	target, err := getIPCAddressPort(cf)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	klog.V(5).Infof("attempting to create grpc agent client connection to: %s", target)
// 	conn, err := grpc.DialContext(ctx, target, opts...)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	klog.V(5).Info("grpc agent client created")
// 	return pb.NewAgentClient(conn), nil
// }
//
// // getIPCAddressPort returns the host and port for connecting to the main agent
// func getIPCAddressPort(cf *config.Config) (string, error) {
// 	ipcAddress, err := cf.GetIPCAddress()
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return net.JoinHostPort(ipcAddress, strconv.Itoa(cf.CmdPort)), nil
// }
