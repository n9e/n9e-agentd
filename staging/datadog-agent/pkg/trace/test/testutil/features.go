package testutil

import (
	"os"

	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/trace/config"
)

// WithFeatures sets the given list of comma-separated features as active and
// returns a function which resets the features to their previous state.
func WithFeatures(feats string) (undo func()) {
	config.SetFeatures(feats)
	return func() {
		config.SetFeatures(os.Getenv("DD_APM_FEATURES"))
	}
}
