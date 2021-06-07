// +build serverless

package processor

import (
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/serverless/aws"
)

// getHostname returns the ARN of the executed function.
func getHostname() string {
	return aws.GetARN()
}
