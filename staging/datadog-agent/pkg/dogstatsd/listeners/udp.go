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
	"time"

	"github.com/DataDog/datadog-agent/pkg/dogstatsd/packets"
	"github.com/DataDog/datadog-agent/pkg/dogstatsd/replay"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/n9e/n9e-agentd/pkg/config"
)

var (
	udpExpvars             = expvar.NewMap("dogstatsd-udp")
	udpPacketReadingErrors = expvar.Int{}
	udpPackets             = expvar.Int{}
	udpBytes               = expvar.Int{}
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
	packetsBuffer   *packets.Buffer
	packetAssembler *packets.Assembler
	buffer          []byte
	trafficCapture  *replay.TrafficCapture // Currently ignored
}

// NewUDPListener returns an idle UDP Statsd listener
func NewUDPListener(packetOut chan packets.Packets, sharedPacketPoolManager *packets.PoolManager, capture *replay.TrafficCapture) (*UDPListener, error) {
	var err error
	var url string

	cf := config.C.Statsd

	if cf.NonLocalTraffic == true {
		// Listen to all network interfaces
		url = fmt.Sprintf(":%d", cf.Port)
	} else {
		url = net.JoinHostPort(config.C.GetBindHost(), strconv.Itoa(cf.Port))
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
	flushTimeout := cf.PacketBufferFlushTimeout.Duration

	buffer := make([]byte, bufferSize)
	packetsBuffer := packets.NewBuffer(uint(packetsBufferSize), flushTimeout, packetOut)
	packetAssembler := packets.NewAssembler(flushTimeout, packetsBuffer, sharedPacketPoolManager, packets.UDP)

	listener := &UDPListener{
		conn:            conn,
		packetsBuffer:   packetsBuffer,
		packetAssembler: packetAssembler,
		buffer:          buffer,
		trafficCapture:  capture,
	}
	log.Debugf("dogstatsd-udp: %s successfully initialized", conn.LocalAddr())
	return listener, nil
}

// Listen runs the intake loop. Should be called in its own goroutine
func (l *UDPListener) Listen() {
	var t1, t2 time.Time
	log.Infof("dogstatsd-udp: starting to listen on %s", l.conn.LocalAddr())
	for {
		n, _, err := l.conn.ReadFrom(l.buffer)
		t1 = time.Now()
		udpPackets.Add(1)

		if err != nil {
			// connection has been closed
			if strings.HasSuffix(err.Error(), " use of closed network connection") {
				return
			}

			log.Errorf("dogstatsd-udp: error reading packet: %v", err)
			udpPacketReadingErrors.Add(1)
			tlmUDPPackets.Inc("error")
		} else {
			tlmUDPPackets.Inc("ok")

			udpBytes.Add(int64(n))
			tlmUDPPacketsBytes.Add(float64(n))

			// packetAssembler merges multiple packets together and sends them when its buffer is full
			l.packetAssembler.AddMessage(l.buffer[:n])
		}

		t2 = time.Now()
		tlmListener.Observe(float64(t2.Sub(t1).Nanoseconds()), "udp")
	}
}

// Stop closes the UDP connection and stops listening
func (l *UDPListener) Stop() {
	l.packetAssembler.Close()
	l.packetsBuffer.Close()
	l.conn.Close()
}
