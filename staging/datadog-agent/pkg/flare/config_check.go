// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package flare

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	"github.com/yubo/apiserver/pkg/cmdcli"

	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/apiserver/response"
)

// configCheckURL contains the Agent API endpoint URL exposing the loaded checks
var configCheckURL string

func GetConfigCheck(w io.Writer, withDebug bool, queries ...string) error {
	var in *api.QueryInput
	if len(queries) > 0 {
		in = &api.QueryInput{Query: queries[0]}
	}

	cr := &response.ConfigCheckResponse{}
	err := common.Client.ApiCall("GET", "/api/v1/config/check", in, nil, cr)
	if err != nil {
		return err
	}

	return getConfigCheck(w, withDebug, cr)
}

// GetClusterAgentConfigCheck proxies GetConfigCheck overidding the URL
func GetClusterAgentConfigCheck(w io.Writer, withDebug bool, queries ...string) error {
	var in *api.QueryInput
	if len(queries) > 0 {
		in = &api.QueryInput{Query: queries[0]}
	}

	cli := common.Client
	cr := &response.ConfigCheckResponse{}
	err := cli.ApiCall("GET", "api/v1/config/check", in, nil,
		cr, cmdcli.WithClient(cli.GetClient("cluster")))
	if err != nil {
		return err
	}
	return getConfigCheck(w, withDebug, cr)
}

// getConfigCheck dump all loaded configurations to the writer
func getConfigCheck(w io.Writer, withDebug bool, cr *response.ConfigCheckResponse) error {
	if w != color.Output {
		color.NoColor = true
	}

	if len(cr.ConfigErrors) > 0 {
		fmt.Fprintln(w, fmt.Sprintf("=== Configuration %s ===", color.RedString("errors")))
		for check, error := range cr.ConfigErrors {
			fmt.Fprintln(w, fmt.Sprintf("\n%s: %s", color.RedString(check), error))
		}
	}

	for _, c := range cr.Configs {
		PrintConfig(w, c)
	}

	if withDebug {
		if len(cr.ResolveWarnings) > 0 {
			fmt.Fprintln(w, fmt.Sprintf("\n=== Resolve %s ===", color.YellowString("warnings")))
			for check, warnings := range cr.ResolveWarnings {
				fmt.Fprintln(w, fmt.Sprintf("\n%s", color.YellowString(check)))
				for _, warning := range warnings {
					fmt.Fprintln(w, fmt.Sprintf("* %s", warning))
				}
			}
		}
		if len(cr.Unresolved) > 0 {
			fmt.Fprintln(w, fmt.Sprintf("\n=== %s Configs ===", color.YellowString("Unresolved")))
			for ids, configs := range cr.Unresolved {
				fmt.Fprintln(w, fmt.Sprintf("\n%s: %s", color.BlueString("Auto-discovery IDs"), color.YellowString(ids)))
				fmt.Fprintln(w, fmt.Sprintf("%s:", color.BlueString("Templates")))
				for _, config := range configs {
					fmt.Fprintln(w, config.String())
				}
			}
		}
	}

	return nil
}

// PrintConfig prints a human-readable representation of a configuration
func PrintConfig(w io.Writer, c integration.Config) {
	if !c.ClusterCheck {
		fmt.Fprintln(w, fmt.Sprintf("\n=== %s check ===", color.GreenString(c.Name)))
	} else {
		fmt.Fprintln(w, fmt.Sprintf("\n=== %s cluster check ===", color.GreenString(c.Name)))
	}

	if c.Provider != "" {
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("Configuration provider"), color.CyanString(c.Provider)))
	} else {
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("Configuration provider"), color.RedString("Unknown provider")))
	}
	if c.Source != "" {
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("Configuration source"), color.CyanString(c.Source)))
	} else {
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("Configuration source"), color.RedString("Unknown configuration source")))
	}
	for _, inst := range c.Instances {
		ID := string(check.BuildID(c.Name, inst, c.InitConfig))
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("Instance ID"), color.CyanString(ID)))
		fmt.Fprint(w, fmt.Sprintf("%s", inst))
		fmt.Fprintln(w, "~")
	}
	if len(c.InitConfig) > 0 {
		fmt.Fprintln(w, fmt.Sprintf("%s:", color.BlueString("Init Config")))
		fmt.Fprintln(w, string(c.InitConfig))
	}
	if len(c.MetricConfig) > 0 {
		fmt.Fprintln(w, fmt.Sprintf("%s:", color.BlueString("Metric Config")))
		fmt.Fprintln(w, string(c.MetricConfig))
	}
	if len(c.LogsConfig) > 0 {
		fmt.Fprintln(w, fmt.Sprintf("%s:", color.BlueString("Log Config")))
		fmt.Fprintln(w, string(c.LogsConfig))
	}
	if len(c.ADIdentifiers) > 0 {
		fmt.Fprintln(w, fmt.Sprintf("%s:", color.BlueString("Auto-discovery IDs")))
		for _, id := range c.ADIdentifiers {
			fmt.Fprintln(w, fmt.Sprintf("* %s", color.CyanString(id)))
		}
	}
	if c.NodeName != "" {
		state := fmt.Sprintf("dispatched to %s", c.NodeName)
		fmt.Fprintln(w, fmt.Sprintf("%s: %s", color.BlueString("State"), color.CyanString(state)))
	}
	fmt.Fprintln(w, "===")
}
