// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package listeners

import (
	"expvar"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/telemetry"
	"k8s.io/klog/v2"

	"github.com/n9e/n9e-agentd/pkg/config"
)

var (
	udpExpvars             = expvar.NewMap("dogstatsd-udp")
	udpPacketReadingErrors = expvar.Int{}
	udpPackets             = expvar.Int{}
	udpBytes               = expvar.Int{}

	tlmUDPPackets = telemetry.NewCounter("statsd", "udp_packets",
		[]string{"state"}, "Dogstatsd UDP packets count")
	tlmUDPPacketsBytes = telemetry.NewCounter("statsd", "udp_packets_bytes",
		nil, "Dogstatsd UDP packets bytes count")
)

func init() {
	udpExpvars.Set("PacketReadingErrors", &udpPacketReadingErrors)
	udpExpvars.Set("Packets", &udpPackets)
	udpExpvars.Set("Bytes", &udpBytes)
}

// UDPListener implements the StatsdListener interface for UDP protocol.
// It listens to a given UDP address and sends back packets ready to be
// processed.
// Origin detection is not implemented for UDP.
type UDPListener struct {
	conn            *net.UDPConn
	packetsBuffer   *packetsBuffer
	packetAssembler *packetAssembler
	buffer          []byte
}

// NewUDPListener returns an idle UDP Statsd listener
func NewUDPListener(packetOut chan Packets, sharedPacketPool *PacketPool) (*UDPListener, error) {
	var err error
	var url string

	cf := config.C.Statsd

	if cf.NonLocalTraffic {
		// Listen to all network interfaces
		url = fmt.Sprintf(":%d", cf.Port)
	} else {
		url = net.JoinHostPort(config.C.GetBindHost(),
			strconv.Itoa(cf.Port))
	}

	addr, err := net.ResolveUDPAddr("udp", url)
	if err != nil {
		return nil, fmt.Errorf("could not resolve udp addr: %s", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("can't listen: %s", err)
	}

	if rcvbuf := cf.SocketRcvbuf; rcvbuf != 0 {
		if err := conn.SetReadBuffer(rcvbuf); err != nil {
			return nil, fmt.Errorf("could not set socket rcvbuf: %s", err)
		}
	}

	bufferSize := cf.BufferSize
	packetsBufferSize := cf.PacketBufferSize
	flushTimeout := cf.PacketBufferFlushTimeout

	buffer := make([]byte, bufferSize)
	packetsBuffer := newPacketsBuffer(uint(packetsBufferSize), flushTimeout, packetOut)
	packetAssembler := newPacketAssembler(flushTimeout, packetsBuffer, sharedPacketPool, UDP)

	listener := &UDPListener{
		conn:            conn,
		packetsBuffer:   packetsBuffer,
		packetAssembler: packetAssembler,
		buffer:          buffer,
	}
	klog.V(5).Infof("statsd-udp: %s successfully initialized", conn.LocalAddr())
	return listener, nil
}

// Listen runs the intake loop. Should be called in its own goroutine
func (l *UDPListener) Listen() {
	klog.Infof("statsd-udp: starting to listen on %s", l.conn.LocalAddr())
	for {
		udpPackets.Add(1)
		n, _, err := l.conn.ReadFrom(l.buffer)
		if err != nil {
			// connection has been closed
			if strings.HasSuffix(err.Error(), " use of closed network connection") {
				return
			}

			klog.Errorf("statsd-udp: error reading packet: %v", err)
			udpPacketReadingErrors.Add(1)
			tlmUDPPackets.Inc("error")
			continue
		}
		tlmUDPPackets.Inc("ok")

		udpBytes.Add(int64(n))
		tlmUDPPacketsBytes.Add(float64(n))

		// packetAssembler merges multiple packets together and sends them when its buffer is full
		l.packetAssembler.addMessage(l.buffer[:n])
	}
}

// Stop closes the UDP connection and stops listening
func (l *UDPListener) Stop() {
	l.packetAssembler.close()
	l.packetsBuffer.close()
	l.conn.Close()
}