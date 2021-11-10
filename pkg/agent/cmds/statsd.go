package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DataDog/datadog-agent/pkg/dogstatsd"
	"github.com/DataDog/datadog-agent/pkg/dogstatsd/replay"
	pb "github.com/DataDog/datadog-agent/pkg/proto/pbgo"
	"github.com/DataDog/datadog-agent/pkg/util/input"
	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/yubo/golib/configer"

	"github.com/spf13/cobra"
)

func newStatsdCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statsd",
		Short: "statsd commands",
	}
	cmd.AddCommand(
		newStatsdCaptureCmd(env),
		newStatsdReplayCmd(env),
		newStatsdStatsCmd(env),
	)

	return cmd
}

func newStatsdCaptureCmd(env *agent.EnvSettings) *cobra.Command {
	input := &api.StatsdCaptureTriggerInput{}

	cmd := &cobra.Command{
		Use:   "statsd-capture",
		Short: "Start a statsd UDS traffic capture",
		RunE: func(cmd *cobra.Command, args []string) error {
			var path string
			err := env.ApiCall("POST", "/api/v1/statsd/capture-trigger", nil, &input, &path)
			if err != nil {
				return err
			}
			fmt.Printf("Capture started, capture file being written to: %s\n", path)
			return nil
		},
	}

	configer.AddFlagsVar(cmd.Flags(), input)

	return cmd
}

func newStatsdReplayCmd(env *agent.EnvSettings) *cobra.Command {
	input := &api.StatsdReplayInput{}

	cmd := &cobra.Command{
		Use:   "statsd-replay",
		Short: "Replay statsd traffic",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statsdReplay(env, input)
		},
	}

	configer.AddFlagsVar(cmd.Flags(), input)

	return cmd
}

func statsdReplay(env *agent.EnvSettings, input *api.StatsdReplayInput) error {
	// setup sig handlers
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigs
		done <- true
	}()

	fmt.Printf("Replaying dogstatsd traffic...\n\n")

	reader, err := replay.NewTrafficCaptureReader(input.ReplayFile, 10)
	if err != nil {
		return err
	}
	defer reader.Close()

	s := config.C.Statsd.Socket
	if s == "" {
		return fmt.Errorf("agent.statsd.socket UNIX socket disabled")
	}

	addr, err := net.ResolveUnixAddr("unixgram", s)
	if err != nil {
		return err
	}

	sk, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(sk)

	err = syscall.SetsockoptInt(sk, syscall.SOL_SOCKET, syscall.SO_SNDBUF, config.C.Statsd.BufferSize)
	if err != nil {
		return err
	}

	dsdSock := os.NewFile(uintptr(sk), "statsd_socket")
	conn, err := net.FileConn(dsdSock)
	if err != nil {
		return err
	}
	defer conn.Close()

	// let's read state before proceeding
	pidmap, state, err := reader.ReadState()
	if err != nil {
		fmt.Printf("Unable to load state from file, tag enrichment will be unavailable for this capture: %v\n", err)
	}

	var resp pb.TaggerStateResponse
	err = env.ApiCall("POST", "/api/v1/statsd/set-tagger-status", nil, &pb.TaggerState{State: state, PidMap: pidmap}, &resp)
	if err != nil {
		fmt.Printf("Unable to load state API error, tag enrichment will be unavailable for this capture: %v\n", err)
	} else if !resp.GetLoaded() {
		fmt.Printf("API refused to set the tagger state, tag enrichment will be unavailable for this capture.\n")
	}

	// enable reading at natural rate
	go reader.Read()

	// wait for go routine to start processing...
	time.Sleep(time.Second)

replay:
	for {
		select {
		case msg := <-reader.Traffic:
			// The cadence is enforced by the reader. The reader will only write to
			// the traffic channel when it estimates the payload should be submitted.
			n, oobn, err := conn.(*net.UnixConn).WriteMsgUnix(
				msg.Payload[:msg.PayloadSize], replay.GetUcredsForPid(msg.Pid), addr)
			if err != nil {
				return err
			}
			fmt.Printf("Sent Payload: %d bytes, and OOB: %d bytes\n", n, oobn)
		case <-reader.Done:
			break replay
		case <-done:
			break replay
		}
	}

	fmt.Println("clearing agent replay states...")

	err = env.ApiCall("POST", "/api/v1/statsd/set-tagger-status", nil, &pb.TaggerState{}, &resp)
	if err != nil {
		fmt.Printf("Unable to load state API error, tag enrichment will be unavailable for this capture: %v\n", err)
	} else if resp.GetLoaded() {
		fmt.Printf("The capture state and pid map have been successfully cleared from the agent\n")
	}

	err = reader.Shutdown()
	if err != nil {
		fmt.Printf("There was an issue shutting down the replay: %v\n", err)
	}

	fmt.Println("replay done")
	return err
}

type statsdStatsInput struct {
	JsonStatus       bool   `flag:"json" description:"print out raw json"`
	PrettyPrintJSON  bool   `flag:"pretty-json" description:"pretty print JSON"`
	DsdStatsFilePath string `flag:"file" description:"Output the dogstatsd-stats command to a file"`
}

func newStatsdStatsCmd(env *agent.EnvSettings) *cobra.Command {
	var input statsdStatsInput

	cmd := &cobra.Command{
		Use:   "statsd-stats",
		Short: "Print basic statistics on the metrics processed by dogstatsd",
		RunE: func(cmd *cobra.Command, args []string) error {
			return statsdStats(env, &input)
		},
	}

	configer.AddFlagsVar(cmd.Flags(), &input)

	return cmd
}

func statsdStats(env *agent.EnvSettings, in *statsdStatsInput) error {

	fmt.Printf("Getting the dogstatsd stats from the agent.\n\n")
	var e error
	var s string

	var r []byte
	err := env.ApiCall("GET", "/api/v1/statsd/stats", nil, nil, &r)
	if err != nil {
		return err
	}

	// The rendering is done in the client so that the agent has less work to do
	if in.PrettyPrintJSON {
		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, r, "", "  ") //nolint:errcheck
		s = prettyJSON.String()
	} else if in.JsonStatus {
		s = string(r)
	} else {
		s, e = dogstatsd.FormatDebugStats(r)
		if e != nil {
			fmt.Printf("Could not format the statistics, the data must be inconsistent. You may want to try the JSON output. Contact the support if you continue having issues.\n")
			return nil
		}
	}

	if in.DsdStatsFilePath == "" {
		fmt.Println(s)
		return nil
	}

	// if the file is already existing, ask for a confirmation.
	if _, err := os.Stat(in.DsdStatsFilePath); err == nil {
		if !input.AskForConfirmation(fmt.Sprintf("'%s' already exists, do you want to overwrite it? [y/N]", in.DsdStatsFilePath)) {
			fmt.Println("Canceling.")
			return nil
		}
	}

	if err := ioutil.WriteFile(in.DsdStatsFilePath, []byte(s), 0644); err != nil {
		fmt.Println("Error while writing the file (is the location writable by the dd-agent user?):", err)
	} else {
		fmt.Println("Dogstatsd stats written in:", in.DsdStatsFilePath)
	}

	return nil
}
