package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/DataDog/datadog-agent/pkg/status"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/yubo/apiserver/pkg/cmdcli"
	"github.com/yubo/golib/configer"
)

type statusInput struct {
	JsonStatus      bool   `flag:"json" description:"print out raw json"`
	PrettyPrintJSON bool   `flag:"pretty-json" description:"pretty print JSON"`
	StatusFilePath  string `flag:"file" description:"Output the status command to a file"`
	component       string
}

func newStatusCmd(env *agent.EnvSettings) *cobra.Command {
	var input statusInput

	cmd := &cobra.Command{
		Use:   "status [component [name]]",
		Short: "Print the current status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return requestStatus(env, &input)
		},
	}

	configer.AddFlags(cmd.Flags(), &input)
	return cmd
}

func newStatusComponentCmd(env *agent.EnvSettings) *cobra.Command {
	var input statusInput

	cmd := &cobra.Command{
		Use:   "component",
		Short: "Print the component status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("a component name must be specified")
			}
			input.component = args[0]
			return componentStatus(env, &input)
		},
	}

	configer.AddFlags(cmd.Flags(), &input)
	return cmd
}

func componentStatus(env *agent.EnvSettings, in *statusInput) error {
	var r []byte
	if err := env.ApiCall("GET",
		fmt.Sprintf("/api/v1/%s/status", in.component),
		nil, nil, &r); err != nil {
		return err
	}

	// The rendering is done in the client so that the agent has less work to do
	var s string
	if in.PrettyPrintJSON {
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, r, "", "  ") //nolint:errcheck
		s = prettyJSON.String()
	} else {
		s = string(r)
	}

	if in.StatusFilePath != "" {
		ioutil.WriteFile(in.StatusFilePath, []byte(s), 0644) //nolint:errcheck
	} else {
		fmt.Println(s)
	}

	return nil
}

func requestStatus(env *agent.EnvSettings, in *statusInput) error {
	var s string

	if !in.PrettyPrintJSON && !in.JsonStatus {
		fmt.Printf("Getting the status from the agent.\n\n")
	}

	var resp map[string]interface{}
	err := env.ApiCall("GET", "/api/v1/status", nil, nil, &resp)
	if err != nil {
		return err
	}
	resp["apmStats"] = getAPMStatus(env)

	r, _ := json.Marshal(resp)

	// The rendering is done in the client so that the agent has less work to do
	if in.PrettyPrintJSON {
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, r, "", "  ") //nolint:errcheck
		s = prettyJSON.String()
	} else if in.JsonStatus {
		s = string(r)
	} else {
		formattedStatus, err := status.FormatStatus(r)
		if err != nil {
			return err
		}
		s = formattedStatus
	}

	if in.StatusFilePath != "" {
		ioutil.WriteFile(in.StatusFilePath, []byte(s), 0644) //nolint:errcheck
	} else {
		fmt.Println(s)
	}

	return nil
}

// getAPMStatus returns a set of key/value pairs summarizing the status of the trace-agent.
// If the status can not be obtained for any reason, the returned map will contain an "error"
// key with an explanation.
func getAPMStatus(env *agent.EnvSettings) map[string]interface{} {
	resp := map[string]interface{}{}

	cli := env.Clients["apm"]
	if cli == nil {
		resp["error"] = fmt.Errorf("unable to get apm client")
		return resp
	}

	err := env.ApiCall("GET", "/debug/vars", nil, nil, &resp, cmdcli.WithClient(cli))
	if err != nil {
		resp["error"] = err.Error()
		return resp
	}

	return resp
}
