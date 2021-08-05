// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package tcp

import (
	"net"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/client"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/logs/types"
)

// AddrToHostPort converts a net.Addr to a (string, int).
func AddrToHostPort(remoteAddr net.Addr) (string, int) {
	switch addr := remoteAddr.(type) {
	case *net.UDPAddr:
		return addr.IP.String(), addr.Port
	case *net.TCPAddr:
		return addr.IP.String(), addr.Port
	}
	return "", 0
}

// AddrToEndPoint creates an EndPoint from an Addr.
func AddrToEndPoint(addr net.Addr) types.Endpoint {
	host, port := AddrToHostPort(addr)
	return types.Endpoint{Host: host, Port: port}
}

// AddrToDestination creates a Destination from an Addr
func AddrToDestination(addr net.Addr, ctx *client.DestinationsContext) *Destination {
	return NewDestination(AddrToEndPoint(addr), true, ctx)
}