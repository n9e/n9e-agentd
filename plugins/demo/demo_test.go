package demo

import (
	"testing"
	"time"

	"github.com/n9e/n9e-agentd/pkg/util/testsender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yubo/golib/util/clock"
)

func TestDemo(t *testing.T) {
	check := &Check{
		count: 8,
		clock: clock.NewFakeClock(time.Unix(0, 0)),
		cos: &cos{
			period: 3600,
			unit:   3600 / 8,
			offset: 1800,
		},
	}

	sender := testsender.NewTestSender(check.ID(), t)
	sender.On("Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	sender.On("Commit").Return()

	assert.Nil(t, check.Run())
	sender.AssertCalled(t, "Gauge", "demo", float64(-1), "", []string{"n:0"})
	sender.AssertCalled(t, "Gauge", "demo", float64(-0.7071067811865477), "", []string{"n:1"})
	sender.AssertCalled(t, "Gauge", "demo", float64(-1.8369701987210272e-16), "", []string{"n:2"})
	sender.AssertCalled(t, "Gauge", "demo", float64(0.7071067811865475), "", []string{"n:3"})
	sender.AssertCalled(t, "Gauge", "demo", float64(1), "", []string{"n:4"})
	sender.AssertCalled(t, "Gauge", "demo", float64(0.7071067811865477), "", []string{"n:5"})
	sender.AssertCalled(t, "Gauge", "demo", float64(3.0616169978683787e-16), "", []string{"n:6"})
	sender.AssertCalled(t, "Gauge", "demo", float64(-0.7071067811865479), "", []string{"n:7"})
	sender.AssertCalled(t, "Commit")

}
