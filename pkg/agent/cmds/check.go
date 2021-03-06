package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/epforwarder"
	"github.com/DataDog/datadog-agent/pkg/flare"
	"github.com/DataDog/datadog-agent/pkg/logs/message"
	"github.com/DataDog/datadog-agent/pkg/metadata"
	"github.com/DataDog/datadog-agent/pkg/serializer"
	"github.com/DataDog/datadog-agent/pkg/status"
	"github.com/DataDog/datadog-agent/pkg/util"
	"github.com/fatih/color"
	"github.com/n9e/n9e-agentd/cmd/agent/common"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/agent/standalone"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/yubo/golib/configer"
	"gopkg.in/yaml.v2"
)

type checkCmd struct {
	env       *agent.EnvSettings
	checkName string

	CheckRate              bool   `flag:"check-rate,r" description:"check rates by running the check twice with a 1sec-pause between the 2 runs"`
	CheckTimes             int    `flag:"check-times" default:"1" description:"number of times to run the check"`
	CheckPause             int    `flag:"pause" default:"0" description:"pause between multiple runs of the check, in milliseconds"`
	LogLevel               string `flag:"log-level,l" description:"set the log level (default 'off') (deprecated, use the env var DD_LOG_LEVEL instead)"`
	CheckDelay             int    `flag:"delay,d" default:"100" description:"delay between running the check and grabbing the metrics in milliseconds"`
	FormatJSON             bool   `flag:"json" description:"format aggregator and check runner output as json"`
	FormatTable            bool   `flag:"table" description:"format aggregator and check runner output as an ascii table"`
	BreakPoint             string `flag:"breakpoint,b" description:"set a breakpoint at a particular line number (Python checks only)"`
	FullSketches           bool   `flag:"full-sketches" description:"output sketches with bins information"`
	SaveFlare              bool   `flag:"flare" description:"save check results to the log dir so it may be reported in a flare"`
	DiscoveryTimeout       uint   `flag:"discovery-timeout" default:"5" description:"max retry duration until Autodiscovery resolves the check template (in seconds)"`
	DiscoveryRetryInterval uint   `flag:"discovery-retry-interval" default:"1" description:"duration between retries until Autodiscovery resolves the check template (in seconds)"`
	ProfileMemory          bool   `flag:"profile-memory,m" description:"run the memory profiler (Python checks only)"`
	ProfileMemoryDir       string `flag:"m-dir" description:"an existing directory in which to store memory profiling data, ignoring clean-up"`
	ProfileMemoryFrames    string `flag:"m-frames" description:"the number of stack frames to consider"`
	ProfileMemoryGC        string `flag:"m-gc" description:"whether or not to run the garbage collector to remove noise"`
	ProfileMemoryCombine   string `flag:"m-combine" description:"whether or not to aggregate over all traceback frames"`
	ProfileMemorySort      string `flag:"m-sort" description:"what to sort by between: lineno | filename | traceback"`
	ProfileMemoryLimit     string `flag:"m-limit" description:"the maximum number of sorted results to show"`
	ProfileMemoryDiff      string `flag:"m-diff" description:"how to order diff results between: absolute | positive"`
	ProfileMemoryFilters   string `flag:"m-filters" description:"comma-separated list of file path glob patterns to filter by"`
	ProfileMemoryUnit      string `flag:"m-unit" description:"the binary unit to represent memory usage (kib, mb, etc.). the default is dynamic"`
	ProfileMemoryVerbose   string `flag:"m-verbose" description:"whether or not to include potentially noisy sources"`

	allConfigs []integration.Config
	cs         []check.Check
	agg        *aggregator.BufferedAggregator
}

// Check returns a cobra command to run checks
func newCheckCmd(env *agent.EnvSettings) *cobra.Command {
	cc := &checkCmd{env: env}

	cmd := &cobra.Command{
		Use:   "check <check_name>",
		Short: "Run the specified check",
		Long:  `Use this to run a specific check with a specific rate`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cc.checkName = args[0]
			return cc.check()
		},
	}

	configer.FlagSet(cmd.Flags(), cc)
	cmd.SetArgs([]string{"checkName"})
	return cmd
}

func (p *checkCmd) check() error {
	hostname, err := util.GetHostname(context.TODO())
	if err != nil {
		fmt.Printf("Cannot get hostname, exiting: %v\n", err)
		return err
	}

	// use the "noop" forwarder because we want the events to be buffered in memory instead of being flushed to the intake
	eventPlatformForwarder := epforwarder.NewNoopEventPlatformForwarder()
	eventPlatformForwarder.Start()

	s := serializer.NewSerializer(common.Forwarder, nil)
	// Initializing the aggregator with a flush interval of 0 (which disable the flush goroutine)
	p.agg = aggregator.InitAggregatorWithFlushInterval(s, eventPlatformForwarder, hostname, 0)
	common.LoadComponents()

	if config.C.InventoriesEnabled {
		metadata.SetupInventoriesExpvar(common.AC, common.Coll)
	}

	if p.DiscoveryRetryInterval > p.DiscoveryTimeout {
		fmt.Println("The discovery retry interval", p.DiscoveryRetryInterval, "is higher than the discovery timeout", p.DiscoveryTimeout)
		fmt.Println("Setting the discovery retry interval to", p.DiscoveryTimeout)
		p.DiscoveryRetryInterval = p.DiscoveryTimeout
	}

	p.allConfigs = waitForConfigs(p.checkName, time.Duration(p.DiscoveryRetryInterval)*time.Second, time.Duration(p.DiscoveryTimeout)*time.Second)

	if err := p.stripJmx(); err != nil {
		return err
	}

	if p.ProfileMemory {
		fn, err := p.setProfileMemory()
		if fn != nil {
			defer fn()
		}
		if err != nil {
			return err
		}
	} else if p.BreakPoint != "" {
		if err := p.setBreakPoint(); err != nil {
			return err
		}
	}

	if err := p.setChecks(); err != nil {
		return err
	}

	if err := p.runChecks(); err != nil {
		return err
	}

	return nil
}

func (p *checkCmd) stripJmx() error {
	// make sure the checks in cs are not JMX checks
	for idx := range p.allConfigs {
		conf := &p.allConfigs[idx]
		if conf.Name != p.checkName {
			continue
		}

		if check.IsJMXConfig(*conf) {
			// we'll mimic the check command behavior with JMXFetch by running
			// it with the JSON reporter and the list_with_metrics command.
			//fmt.Println("Please consider using the 'jmx' command instead of 'check jmx'")
			selectedChecks := []string{p.checkName}
			if p.CheckRate {
				if err := standalone.ExecJmxListWithRateMetricsJSON(p.env, selectedChecks, p.LogLevel); err != nil {
					return fmt.Errorf("while running the jmx check: %v", err)
				}
			} else {
				if err := standalone.ExecJmxListWithMetricsJSON(p.env, selectedChecks, p.LogLevel); err != nil {
					return fmt.Errorf("while running the jmx check: %v", err)
				}
			}

			instances := []integration.Data{}

			// Retain only non-JMX instances for later
			for _, instance := range conf.Instances {
				if check.IsJMXInstance(conf.Name, instance, conf.InitConfig) {
					continue
				}
				instances = append(instances, instance)
			}

			if len(instances) == 0 {
				return fmt.Errorf("All instances of '%s' are JMXFetch instances, and have completed running\n", p.checkName)
				//return nil
			}

			conf.Instances = instances
		}
	}

	return nil
}

func (p *checkCmd) setProfileMemory() (fn func(), err error) {
	// If no directory is specified, make a temporary one
	if p.ProfileMemoryDir == "" {
		if p.ProfileMemoryDir, err = ioutil.TempDir("", "datadog-agent-memory-profiler"); err != nil {
			return
		}

		fn = func() {
			cleanupErr := os.RemoveAll(p.ProfileMemoryDir)
			if cleanupErr != nil {
				fmt.Printf("%s\n", cleanupErr)
			}
		}
	}

	for idx := range p.allConfigs {
		conf := &p.allConfigs[idx]
		if conf.Name != p.checkName {
			continue
		}

		var data map[string]interface{}

		if err = yaml.Unmarshal(conf.InitConfig, &data); err != nil {
			return
		}

		if data == nil {
			data = make(map[string]interface{})
		}

		data["profile_memory"] = p.ProfileMemoryDir
		if err = p.populateMemoryProfileConfig(data); err != nil {
			return
		}

		y, _ := yaml.Marshal(data)
		conf.InitConfig = y

		break
	}

	return
}

func (p *checkCmd) setBreakPoint() error {
	breakPointLine, err := strconv.Atoi(p.BreakPoint)
	if err != nil {
		fmt.Printf("breakpoint must be an integer\n")
		return err
	}

	for idx := range p.allConfigs {
		conf := &p.allConfigs[idx]
		if conf.Name != p.checkName {
			continue
		}

		var data map[string]interface{}

		err = yaml.Unmarshal(conf.InitConfig, &data)
		if err != nil {
			return err
		}

		if data == nil {
			data = make(map[string]interface{})
		}

		data["set_breakpoint"] = breakPointLine

		y, _ := yaml.Marshal(data)
		conf.InitConfig = y

		break
	}

	return nil
}

func (p *checkCmd) setChecks() error {
	p.cs = collector.GetChecksByNameForConfigs(p.checkName, p.allConfigs)

	if len(p.cs) > 1 {
		fmt.Println("Multiple check instances found, running each of them")
		return nil
	}

	if len(p.cs) > 0 {
		return nil
	}

	// something happened while getting the check(s), display some info.
	//if len(p.cs) == 0 {
	for check, error := range autodiscovery.GetConfigErrors() {
		if p.checkName == check {
			fmt.Fprintln(color.Output, fmt.Sprintf("\n%s: invalid config for %s: %s", color.RedString("Error"), color.YellowString(check), error))
		}
	}
	for check, errors := range collector.GetLoaderErrors() {
		if p.checkName == check {
			fmt.Fprintln(color.Output, fmt.Sprintf("\n%s: could not load %s:", color.RedString("Error"), color.YellowString(p.checkName)))
			for loader, error := range errors {
				fmt.Fprintln(color.Output, fmt.Sprintf("* %s: %s", color.YellowString(loader), error))
			}
		}
	}
	for check, warnings := range autodiscovery.GetResolveWarnings() {
		if p.checkName == check {
			fmt.Fprintln(color.Output, fmt.Sprintf("\n%s: could not resolve %s config:", color.YellowString("Warning"), color.YellowString(check)))
			for _, warning := range warnings {
				fmt.Fprintln(color.Output, fmt.Sprintf("* %s", warning))
			}
		}
	}
	return fmt.Errorf("no valid check found")
	//}
}

func (p *checkCmd) runChecks() error {
	var checkFileOutput bytes.Buffer
	var instancesData []interface{}
	for _, c := range p.cs {
		// <<--
		s := p.runCheck(c)

		// Sleep for a while to allow the aggregator to finish ingesting all the metrics/events/sc
		time.Sleep(time.Duration(p.CheckDelay) * time.Millisecond)

		if p.FormatJSON {
			aggregatorData := getMetricsData(p.agg)
			var collectorData map[string]interface{}

			collectorJSON, _ := status.GetCheckStatusJSON(c, s)
			if err := json.Unmarshal(collectorJSON, &collectorData); err != nil {
				return err
			}

			checkRuns := collectorData["runnerStats"].(map[string]interface{})["Checks"].(map[string]interface{})[p.checkName].(map[string]interface{})

			// There is only one checkID per run so we'll just access that
			var runnerData map[string]interface{}
			for _, checkIDData := range checkRuns {
				runnerData = checkIDData.(map[string]interface{})
				break
			}

			instanceData := map[string]interface{}{
				"aggregator":  aggregatorData,
				"runner":      runnerData,
				"inventories": collectorData["inventories"],
			}
			instancesData = append(instancesData, instanceData)
			continue
		}

		if p.ProfileMemory {
			// Every instance will create its own directory
			instanceID := strings.SplitN(string(c.ID()), ":", 2)[1]
			// Colons can't be part of Windows file paths
			instanceID = strings.Replace(instanceID, ":", "_", -1)
			profileDataDir := filepath.Join(p.ProfileMemoryDir, p.checkName, instanceID)

			snapshotDir := filepath.Join(profileDataDir, "snapshots")
			if _, err := os.Stat(snapshotDir); !os.IsNotExist(err) {
				snapshots, err := ioutil.ReadDir(snapshotDir)
				if err != nil {
					return err
				}

				numSnapshots := len(snapshots)
				if numSnapshots > 0 {
					lastSnapshot := snapshots[numSnapshots-1]
					snapshotContents, err := ioutil.ReadFile(filepath.Join(snapshotDir, lastSnapshot.Name()))
					if err != nil {
						return err
					}

					color.HiWhite(string(snapshotContents))
				} else {
					return fmt.Errorf("no snapshots found in %s", snapshotDir)
				}
			} else {
				return fmt.Errorf("no snapshot data found in %s", profileDataDir)
			}

			diffDir := filepath.Join(profileDataDir, "diffs")
			if _, err := os.Stat(diffDir); !os.IsNotExist(err) {
				diffs, err := ioutil.ReadDir(diffDir)
				if err != nil {
					return err
				}

				numDiffs := len(diffs)
				if numDiffs > 0 {
					lastDiff := diffs[numDiffs-1]
					diffContents, err := ioutil.ReadFile(filepath.Join(diffDir, lastDiff.Name()))
					if err != nil {
						return err
					}

					color.HiCyan(fmt.Sprintf("\n%s\n\n", strings.Repeat("=", 50)))
					color.HiWhite(string(diffContents))
				} else {
					return fmt.Errorf("no diffs found in %s", diffDir)
				}
			} else if !p.singleCheckRun() {
				return fmt.Errorf("no diff data found in %s", profileDataDir)
			}
			continue
		}

		p.printMetrics(&checkFileOutput)
		checkStatus, _ := status.GetCheckStatus(c, s)
		statusString := string(checkStatus)
		fmt.Println(statusString)
		checkFileOutput.WriteString(statusString + "\n")
	}

	//if runtime.GOOS == "windows" {
	//	standalone.PrintWindowsUserWarning("check")
	//}

	if p.FormatJSON {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("JSON")))
		checkFileOutput.WriteString("=== JSON ===\n")

		instancesJSON, _ := json.MarshalIndent(instancesData, "", "  ")
		instanceJSONString := string(instancesJSON)

		fmt.Println(instanceJSONString)
		checkFileOutput.WriteString(instanceJSONString + "\n")
	} else if p.singleCheckRun() {
		if p.ProfileMemory {
			color.Yellow("Check has run only once, to collect diff data run the check multiple times with the --check-times flag.")
		} else {
			color.Yellow("Check has run only once, if some metrics are missing you can try again with --check-rate to see any other metric if available.")
		}
	}

	//if warnings != nil && warnings.TraceMallocEnabledWithPy2 {
	//	return errors.New("tracemalloc is enabled but unavailable with python version 2")
	//}

	if p.SaveFlare {
		p.writeCheckToFile(p.checkName, &checkFileOutput)
	}

	return nil
}

func (p *checkCmd) runCheck(c check.Check) *check.Stats {
	s := check.NewStats(c)
	times := p.CheckTimes
	pause := p.CheckPause
	if p.CheckRate {
		if p.CheckTimes > 2 {
			color.Yellow("The check-rate option is overriding check-times to 2")
		}
		if pause > 0 {
			color.Yellow("The check-rate option is overriding pause to 1000ms")
		}
		times = 2
		pause = 1000
	}
	for i := 0; i < times; i++ {
		t0 := time.Now()
		err := c.Run()
		warnings := c.GetWarnings()
		sStats, _ := c.GetSenderStats()
		s.Add(time.Since(t0), err, warnings, sStats)
		if pause > 0 && i < times-1 {
			time.Sleep(time.Duration(pause) * time.Millisecond)
		}
	}

	return s
}

func (p *checkCmd) printMetrics(checkFileOutput *bytes.Buffer) {
	series, sketches := p.agg.GetSeriesAndSketches(time.Now())
	if len(series) != 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Series")))

		if p.FormatTable {
			headers, data := series.MarshalStrings()
			var buffer bytes.Buffer

			// plain table with no borders
			table := tablewriter.NewWriter(&buffer)
			table.SetHeader(headers)
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(true)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")
			table.SetHeaderLine(false)
			table.SetBorder(false)
			table.SetTablePadding("\t")

			table.AppendBulk(data)
			table.Render()
			fmt.Println(buffer.String())
			checkFileOutput.WriteString(buffer.String() + "\n")
		} else {
			j, _ := json.MarshalIndent(series, "", "  ")
			fmt.Println(string(j))
			checkFileOutput.WriteString(string(j) + "\n")
		}
	}
	if len(sketches) != 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Sketches")))
		j, _ := json.MarshalIndent(sketches, "", "  ")
		fmt.Println(string(j))
		checkFileOutput.WriteString(string(j) + "\n")
	}

	serviceChecks := p.agg.GetServiceChecks()
	if len(serviceChecks) != 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Service Checks")))

		if p.FormatTable {
			headers, data := serviceChecks.MarshalStrings()
			var buffer bytes.Buffer

			// plain table with no borders
			table := tablewriter.NewWriter(&buffer)
			table.SetHeader(headers)
			table.SetAutoWrapText(false)
			table.SetAutoFormatHeaders(true)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetRowSeparator("")
			table.SetHeaderLine(false)
			table.SetBorder(false)
			table.SetTablePadding("\t")

			table.AppendBulk(data)
			table.Render()
			fmt.Println(buffer.String())
			checkFileOutput.WriteString(buffer.String() + "\n")
		} else {
			j, _ := json.MarshalIndent(serviceChecks, "", "  ")
			fmt.Println(string(j))
			checkFileOutput.WriteString(string(j) + "\n")
		}
	}

	events := p.agg.GetEvents()
	if len(events) != 0 {
		fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString("Events")))
		checkFileOutput.WriteString("=== Events ===\n")
		j, _ := json.MarshalIndent(events, "", "  ")
		fmt.Println(string(j))
		checkFileOutput.WriteString(string(j) + "\n")
	}

	for k, v := range toDebugEpEvents(p.agg.GetEventPlatformEvents()) {
		if len(v) > 0 {
			if translated, ok := check.EventPlatformNameTranslations[k]; ok {
				k = translated
			}
			fmt.Fprintln(color.Output, fmt.Sprintf("=== %s ===", color.BlueString(k)))
			checkFileOutput.WriteString(fmt.Sprintf("=== %s ===\n", k))
			j, _ := json.MarshalIndent(v, "", "  ")
			fmt.Println(string(j))
			checkFileOutput.WriteString(string(j) + "\n")
		}
	}
}

func (p *checkCmd) writeCheckToFile(checkName string, checkFileOutput *bytes.Buffer) {
	flareDir := config.C.CheckFlareDir
	_ = os.Mkdir(flareDir, os.ModeDir)

	// Windows cannot accept ":" in file names
	filenameSafeTimeStamp := strings.ReplaceAll(time.Now().UTC().Format(time.RFC3339), ":", "-")
	flarePath := filepath.Join(flareDir, "check_"+checkName+"_"+filenameSafeTimeStamp+".log")

	w, err := flare.NewRedactingWriter(flarePath, os.ModePerm, true)
	if err != nil {
		fmt.Println("Error while writing the check file:", err)
		return
	}
	defer w.Close()

	_, err = w.Write(checkFileOutput.Bytes())

	if err != nil {
		fmt.Println("Error while writing the check file (is the location writable by the dd-agent user?):", err)
	} else {
		fmt.Println("check written to:", flarePath)
	}
}

type eventPlatformDebugEvent struct {
	RawEvent          string `json:",omitempty"`
	EventType         string
	UnmarshalledEvent map[string]interface{} `json:",omitempty"`
}

// toDebugEpEvents transforms the raw event platform messages to eventPlatformDebugEvents which are better for json formatting
func toDebugEpEvents(events map[string][]*message.Message) map[string][]eventPlatformDebugEvent {
	result := make(map[string][]eventPlatformDebugEvent)
	for eventType, messages := range events {
		var events []eventPlatformDebugEvent
		for _, m := range messages {
			e := eventPlatformDebugEvent{EventType: eventType, RawEvent: string(m.Content)}
			err := json.Unmarshal([]byte(e.RawEvent), &e.UnmarshalledEvent)
			if err == nil {
				e.RawEvent = ""
			}
			events = append(events, e)
		}
		result[eventType] = events
	}
	return result
}

func getMetricsData(agg *aggregator.BufferedAggregator) map[string]interface{} {
	aggData := make(map[string]interface{})

	series, sketches := agg.GetSeriesAndSketches(time.Now())
	if len(series) != 0 {
		// Workaround to get the raw sequence of metrics, see:
		// https://github.com/DataDog/datadog-agent/blob/b2d9527ec0ec0eba1a7ae64585df443c5b761610/pkg/metrics/series.go#L109-L122
		var data map[string]interface{}
		sj, _ := json.Marshal(series)
		json.Unmarshal(sj, &data) //nolint:errcheck

		aggData["metrics"] = data["series"]
	}
	if len(sketches) != 0 {
		aggData["sketches"] = sketches
	}

	serviceChecks := agg.GetServiceChecks()
	if len(serviceChecks) != 0 {
		aggData["service_checks"] = serviceChecks
	}

	events := agg.GetEvents()
	if len(events) != 0 {
		aggData["events"] = events
	}

	for k, v := range toDebugEpEvents(agg.GetEventPlatformEvents()) {
		aggData[k] = v
	}

	return aggData
}

func (p *checkCmd) singleCheckRun() bool {
	return p.CheckRate == false && p.CheckTimes < 2
}

func createHiddenStringFlag(cmd *cobra.Command, p *string, name string, value string, usage string) {
	cmd.Flags().StringVar(p, name, value, usage)
	cmd.Flags().MarkHidden(name) //nolint:errcheck
}

func (p *checkCmd) populateMemoryProfileConfig(initConfig map[string]interface{}) error {
	if p.ProfileMemoryFrames != "" {
		profileMemoryFrames, err := strconv.Atoi(p.ProfileMemoryFrames)
		if err != nil {
			return fmt.Errorf("--m-frames must be an integer")
		}
		initConfig["profile_memory_frames"] = profileMemoryFrames
	}

	if p.ProfileMemoryGC != "" {
		profileMemoryGC, err := strconv.Atoi(p.ProfileMemoryGC)
		if err != nil {
			return fmt.Errorf("--m-gc must be an integer")
		}

		initConfig["profile_memory_gc"] = profileMemoryGC
	}

	if p.ProfileMemoryCombine != "" {
		profileMemoryCombine, err := strconv.Atoi(p.ProfileMemoryCombine)
		if err != nil {
			return fmt.Errorf("--m-combine must be an integer")
		}

		if profileMemoryCombine != 0 && p.ProfileMemorySort == "traceback" {
			return fmt.Errorf("--m-combine cannot be sorted (--m-sort) by traceback")
		}

		initConfig["profile_memory_combine"] = profileMemoryCombine
	}

	if p.ProfileMemorySort != "" {
		if p.ProfileMemorySort != "lineno" && p.ProfileMemorySort != "filename" && p.ProfileMemorySort != "traceback" {
			return fmt.Errorf("--m-sort must one of: lineno | filename | traceback")
		}
		initConfig["profile_memory_sort"] = p.ProfileMemorySort
	}

	if p.ProfileMemoryLimit != "" {
		profileMemoryLimit, err := strconv.Atoi(p.ProfileMemoryLimit)
		if err != nil {
			return fmt.Errorf("--m-limit must be an integer")
		}
		initConfig["profile_memory_limit"] = profileMemoryLimit
	}

	if p.ProfileMemoryDiff != "" {
		if p.ProfileMemoryDiff != "absolute" && p.ProfileMemoryDiff != "positive" {
			return fmt.Errorf("--m-diff must one of: absolute | positive")
		}
		initConfig["profile_memory_diff"] = p.ProfileMemoryDiff
	}

	if p.ProfileMemoryFilters != "" {
		initConfig["profile_memory_filters"] = p.ProfileMemoryFilters
	}

	if p.ProfileMemoryUnit != "" {
		initConfig["profile_memory_unit"] = p.ProfileMemoryUnit
	}

	if p.ProfileMemoryVerbose != "" {
		profileMemoryVerbose, err := strconv.Atoi(p.ProfileMemoryVerbose)
		if err != nil {
			return fmt.Errorf("--m-verbose must be an integer")
		}
		initConfig["profile_memory_verbose"] = profileMemoryVerbose
	}

	return nil
}

// containsCheck returns true if at least one config corresponds to the check name.
func containsCheck(checkName string, configs []integration.Config) bool {
	for _, cfg := range configs {
		if cfg.Name == checkName {
			return true
		}
	}

	return false
}

// waitForConfigs retries the collection of Autodiscovery configs until the check is found or the timeout is reached.
// Autodiscovery listeners run asynchronously, AC.GetAllConfigs() can fail at the beginning to resolve templated configs
// depending on non-deterministic factors (system load, network latency, active Autodiscovery listeners and their configurations).
// This function improves the resiliency of the check command.
// Note: If the check corresponds to a non-template configuration it should be found on the first try and fast-returned.
func waitForConfigs(checkName string, retryInterval, timeout time.Duration) []integration.Config {
	allConfigs := common.AC.GetAllConfigs()
	if containsCheck(checkName, allConfigs) {
		return allConfigs
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	retryTicker := time.NewTicker(retryInterval)
	defer retryTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return allConfigs
		case <-retryTicker.C:
			allConfigs = common.AC.GetAllConfigs()
			if containsCheck(checkName, allConfigs) {
				return allConfigs
			}
		}
	}
}

// TODO: re-enable when the API endpoint is implemented
func newListchecksCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "list-checks",
		Short: "Query the agent for the list of checks running",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(env.Out, "Checks: ")
			env.ApiCall("GET", "/api/v1/checks", nil, nil, env.Out)
			fmt.Fprintf(env.Out, "\n")
			return nil
		},
	}
}

func newReloadChecksCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "reload-check <check_name>",
		Short: "Reload a running check",
		RunE: func(cmd *cobra.Command, args []string) error {

			var checkName string
			if len(args) != 0 {
				checkName = args[0]
			} else {
				return fmt.Errorf("missing arguments")
			}

			fmt.Fprintf(env.Out, "Reload check %s: ", checkName)
			env.ApiCall("POST", fmt.Sprintf("/api/v1/checks/%s/reload", checkName),
				nil, nil, env.Out)
			fmt.Fprintf(env.Out, "\n")
			return nil
		},
	}
}
