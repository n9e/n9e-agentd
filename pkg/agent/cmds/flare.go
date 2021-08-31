package cmds

import (
	"fmt"
	"os"

	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/util/input"
	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/configer"
)

type flareCmd struct {
	env                  *agent.EnvSettings
	caseID               string
	CustomerEmail        string `flag:"email,e" description:"Your email"`
	Autoconfirm          bool   `flag:"send,s" description:"Automatically send flare (don't prompt for confirmation)"`
	ForceLocal           bool   `flag:"local,l" description:"Force the creation of the flare by the command line instead of the agent process (useful when running in a containerized env)"`
	Profiling            int    `flag:"profile,p" default:"-1" description:"Add performance profiling data to the flare. It will collect a heap profile and a CPU profile for the amount of seconds passed to the flag, with a minimum of 30s"`
	ProfileMutex         bool   `flag:"profile-mutex,M" description:"Add mutex profile to the performance data in the flare"`
	ProfileMutexFraction int    `flag:"profile-mutex-fraction" default:"100" description:"Set the fraction of mutex contention events that are reported in the mutex profile"`
	ProfileBlocking      bool   `flag:"profile-blocking,B" description:"Add gorouting blocking profile to the performance data in the flare"`
	ProfileBlockingRate  int    `flag:"profile-blocking-rate" default:"10000" description:"Set the fraction of goroutine blocking events that are reported in the blocking profile"`
}

func newFlareCmd(env *agent.EnvSettings) *cobra.Command {
	var in flareCmd
	cmd := &cobra.Command{
		Use:   "flare [caseID]",
		Short: "Collect a flare and send it to Datadog",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				in.caseID = args[0]
			}
			if in.CustomerEmail == "" {
				var err error
				in.CustomerEmail, err = input.AskForEmail()
				if err != nil {
					fmt.Println("Error reading email, please retry or contact support")
					return err
				}
			}

			in.env = env

			return in.makeFlare()
		},
	}

	configer.AddFlags(cmd.Flags(), &in)
	cmd.SetArgs([]string{"caseID"})
	return cmd

}

func (p *flareCmd) makeFlare() error {
	logFile := config.C.LogFile
	jmxLogFile := config.C.Jmx.LogFile
	logFiles := []string{logFile, jmxLogFile}
	var (
		profile flare.ProfileData
		err     error
	)
	if p.Profiling >= 30 {
		if err := p.readProfileData(&profile); err != nil {
			fmt.Fprintln(color.Output, color.RedString(fmt.Sprintf("Could not collect performance profile: %s", err)))
			return err
		}
	} else if p.Profiling != -1 {
		fmt.Fprintln(color.Output, color.RedString(fmt.Sprintf("Invalid value for profiling: %d. Please enter an integer of at least 30.", p.Profiling)))
		return err
	}

	var filePath string
	if p.ForceLocal {
		filePath, err = p.createArchive(logFiles, profile)
	} else {
		filePath, err = p.requestArchive(logFiles, profile)
	}

	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); err != nil {
		fmt.Fprintln(color.Output, color.RedString(fmt.Sprintf("The flare zipfile \"%s\" does not exist.", filePath)))
		fmt.Fprintln(color.Output, color.RedString("If the agent running in a different container try the '--local' option to generate the flare locally"))
		return err
	}

	fmt.Fprintln(color.Output, fmt.Sprintf("%s is going to be uploaded to Datadog", color.YellowString(filePath)))
	if !p.Autoconfirm {
		confirmation := input.AskForConfirmation("Are you sure you want to upload a flare? [y/N]")
		if !confirmation {
			fmt.Fprintln(color.Output, fmt.Sprintf("Aborting. (You can still use %s)", color.YellowString(filePath)))
			return nil
		}
	}

	response, e := flare.SendFlare(filePath, p.caseID, p.CustomerEmail)
	fmt.Println(response)
	if e != nil {
		return e
	}
	return nil
}

func (p *flareCmd) readProfileData(pdata *flare.ProfileData) error {
	prevSettings, err := p.setRuntimeProfilingSettings()
	if err != nil {
		return err
	}
	defer p.resetRuntimeProfilingSettings(prevSettings)

	cli := p.env.Clients["exporter"]
	if cli == nil {
		return fmt.Errorf("unable to get pprof client")
	}

	fmt.Fprintln(color.Output, color.BlueString("Getting a %ds profile snapshot from core.", p.Profiling))

	// ignore apm
	return flare.CreatePerformanceProfile(p.env, cli, "core", p.Profiling, pdata)
}

func (p *flareCmd) requestArchive(logFiles []string, pdata flare.ProfileData) (string, error) {
	fmt.Fprintln(color.Output, color.BlueString("Asking the agent to build the flare archive."))

	r := []byte{}
	e := p.env.ApiCall("POST", "/api/v1/flare", nil, pdata, &r)
	if e != nil {
		if r != nil && string(r) != "" {
			fmt.Fprintln(color.Output, fmt.Sprintf("The agent ran into an error while making the flare: %s", color.RedString(string(r))))
		} else {
			fmt.Fprintln(color.Output, color.RedString("The agent was unable to make the flare. (is it running?)"))
		}
		return p.createArchive(logFiles, pdata)
	}
	return string(r), nil
}

func (p *flareCmd) createArchive(logFiles []string, pdata flare.ProfileData) (string, error) {
	fmt.Fprintln(color.Output, color.YellowString("Initiating flare locally."))
	filePath, e := flare.CreateArchive(true, config.C.DistPath, config.C.PyChecksPath, logFiles, pdata)
	if e != nil {
		fmt.Printf("The flare zipfile failed to be created: %s\n", e)
		return "", e
	}
	return filePath, nil
}

func (p *flareCmd) setRuntimeProfilingSettings() (map[string]interface{}, error) {
	prev := make(map[string]interface{})
	if p.ProfileMutex && p.ProfileMutexFraction > 0 {
		old, err := p.setRuntimeSetting("runtime_mutex_profile_fraction", p.ProfileMutexFraction)
		if err != nil {
			return nil, err
		}
		prev["runtime_mutex_profile_fraction"] = old
	}
	if p.ProfileBlocking && p.ProfileBlockingRate > 0 {
		old, err := p.setRuntimeSetting("runtime_block_profile_rate", p.ProfileBlockingRate)
		if err != nil {
			return nil, err
		}
		prev["runtime_block_profile_rate"] = old
	}
	return prev, nil
}

func (p *flareCmd) setRuntimeSetting(name string, new int) (interface{}, error) {
	fmt.Fprintln(color.Output, color.BlueString("Setting %s to %v", name, new))
	c := agent.NewSettingsClient(p.env)

	oldVal, err := c.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get current value of %s: %v", name, err)
	}

	if _, err := c.Set(name, fmt.Sprint(new)); err != nil {
		return nil, fmt.Errorf("failed to set %s to %v: %v", name, new, err)
	}

	return oldVal, nil
}

func (p *flareCmd) resetRuntimeProfilingSettings(prev map[string]interface{}) {
	if len(prev) == 0 {
		return
	}

	c := agent.NewSettingsClient(p.env)

	for name, value := range prev {
		fmt.Fprintln(color.Output, color.BlueString("Restoring %s to %v", name, value))
		if _, err := c.Set(name, fmt.Sprint(value)); err != nil {
			fmt.Fprintln(color.Output, color.RedString("Failed to restore previous value of %s: %v", name, err))
		}
	}
}
