package port

import (
	"fmt"
	"net"
	"testing"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/autodiscovery/validation"
	"github.com/stretchr/testify/assert"
)

func TestCheckTCP(t *testing.T) {
	check := new(Check)

	l := listenTCP()
	err := check.Configure([]byte(`protocol: tcp`), nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	check.config.addrs = []string{l.Addr().String()}

	ok := check.check()
	assert.Equal(t, true, ok)

	l.Close()
	ok = check.check()
	assert.Equal(t, false, ok)
}

func TestCheckUDP(t *testing.T) {
	check := new(Check)

	addr, l := listenUDP()
	err := check.Configure([]byte(`protocol: udp`), nil, "test")
	if err != nil {
		t.Fatal(err)
	}
	check.config.addrs = []string{addr.String()}

	ok := check.check()
	assert.Equal(t, true, ok)

	l.Close()
	ok = check.check()
	assert.Equal(t, false, ok)

}

func listenUDP() (*net.UDPAddr, *net.UDPConn) {
	for i := 10000; i < 20000; i++ {
		addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("172.0.0.1:%d", i))
		l, err := net.ListenUDP("udp", addr)
		if err != nil {
			addr, _ = net.ResolveUDPAddr("udp6", fmt.Sprintf("[::1]:%d", i))
			if l, err = net.ListenUDP("udp6", addr); err != nil {
				continue
			}
		}
		return addr, l

	}
	panic("failed to listen UDP port")

}

func listenTCP() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("failed to listen on a port: %v", err))
		}
	}
	return l
}

func TestConfig(t *testing.T) {
	cases := []struct {
		config string
		ok     bool
	}{
		{"", false},
		{"{}", false},
		{`
{
  "initConfig": { "tiemout": 3 },
  "instances": [{
    "minCollectionInterval": 10,
    "tags": ["a:1", "b:2"],
    "port": 123,
    "protocol": "tcp"
  }]
}`, true},
	}

	for _, c := range cases {
		err := validation.ValidateJSONConfig(checkName, []byte(c.config))
		assert.Equal(t, c.ok, err == nil)
	}
}
