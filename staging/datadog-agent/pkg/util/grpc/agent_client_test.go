package grpc

import (
	"context"
	"os"
	"testing"
	"time"

	"k8s.io/klog/v2"
	"github.com/cihub/seelog"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.SetupLogger(seelog.Default, "trace")
	os.Exit(m.Run())
}

func TestGetDDAgentClientTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := GetDDAgentClient(ctx)
	assert.Equal(t, context.DeadlineExceeded, err)
}
