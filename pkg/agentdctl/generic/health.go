package generic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/agentdctl"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/spf13/cobra"
)

func healthCmd(env *agentdctl.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "health",
		Short:        "Print the current agent health",
		Long:         ``,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			health(env)
			return nil
		},
	}
	return cmd
}

func health(env *agentdctl.EnvSettings) error {
	s := &api.HealthStatus{}
	if err := env.ApiCall("GET", "cmd", "/api/status/health", nil, s); err != nil {
		return err
	}

	sort.Strings(s.Unhealthy)
	sort.Strings(s.Healthy)

	statusString := color.GreenString("PASS")
	if len(s.Unhealthy) > 0 {
		statusString = color.RedString("FAIL")
	}
	fmt.Fprintln(color.Output, fmt.Sprintf("Agent health: %s", statusString))

	if len(s.Healthy) > 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s healthy components ===", color.GreenString(strconv.Itoa(len(s.Healthy)))))
		fmt.Fprintln(color.Output, strings.Join(s.Healthy, ", "))
	}
	if len(s.Unhealthy) > 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s unhealthy components ===", color.RedString(strconv.Itoa(len(s.Unhealthy)))))
		fmt.Fprintln(color.Output, strings.Join(s.Unhealthy, ", "))
		return fmt.Errorf("found %d unhealthy components", len(s.Unhealthy))
	}

	return nil
}
