//go:build secrets
// +build secrets

package cmds

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"

	s "github.com/DataDog/datadog-agent/pkg/secrets"
)

const (
	maxSecretFileSize = 8192
)

func newSecretCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Print information about decrypted secrets in configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return env.ApiCallDone("GET", "/api/v1/secrets", nil, nil, &s.SecretInfo{})
		},
	}
	return cmd
}

func newSecretHelpCmd(env *agent.EnvSettings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret-helper",
		Short: "Secret management backend helper",
	}

	cmd.AddCommand(newReadSecretsCmd(env))

	return cmd
}

// ReadSecretsCmd implements reading secrets from a directory/volume mount
func newReadSecretsCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "read",
		Short: "Read secret from a directory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return readSecrets(os.Stdin, os.Stdout, args[0])
		},
	}
}

type secretsRequest struct {
	Version string   `json:"version"`
	Secrets []string `json:"secrets"`
}

// ReadSecrets implements a secrets reader from a directory/mount
func readSecrets(r io.Reader, w io.Writer, dir string) error {
	in, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	var request secretsRequest
	err = json.Unmarshal(in, &request)
	if err != nil {
		return errors.New("failed to unmarshal json input")
	}

	version := splitVersion(request.Version)
	compatVersion := splitVersion(s.PayloadVersion)
	if version[0] != compatVersion[0] {
		return fmt.Errorf("incompatible protocol version %q", request.Version)
	}

	if len(request.Secrets) == 0 {
		return errors.New("no secrets listed in input")
	}

	response := map[string]s.Secret{}
	for _, name := range request.Secrets {
		response[name] = readSecret(filepath.Join(dir, name))
	}

	out, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = w.Write(out)
	return err
}

func splitVersion(ver string) []string {
	return strings.SplitN(ver, ".", 2)
}

func readSecret(path string) s.Secret {
	value, err := readSecretFile(path)
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	return s.Secret{Value: value, ErrorMsg: errMsg}
}

func readSecretFile(path string) (string, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("secret does not exist")
		}
		return "", err
	}

	if fi.Mode()&os.ModeSymlink != 0 {
		// Ensure that the symlink is in the same dir
		target, err := os.Readlink(path)
		if err != nil {
			return "", fmt.Errorf("failed to read symlink target: %v", err)
		}

		dir := filepath.Dir(path)
		if !filepath.IsAbs(target) {
			target, err = filepath.Abs(filepath.Join(dir, target))
			if err != nil {
				return "", fmt.Errorf("failed to resolve symlink absolute path: %v", err)
			}
		}

		dirAbs, err := filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve absolute path of directory: %v", err)
		}

		if !filepath.HasPrefix(target, dirAbs) {
			return "", fmt.Errorf("not following symlink %q outside of %q", target, dir)
		}
	}
	fi, err = os.Stat(path)
	if err != nil {
		return "", err
	}

	if fi.Size() > maxSecretFileSize {
		return "", errors.New("secret exceeds max allowed size")
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func init() {
	agent.RegisterCmd([]agent.CmdOps{
		{CmdFactory: newSecretCmd, GroupNum: agent.CMD_G_GENERIC},
		{CmdFactory: newSecretHelpCmd, GroupNum: agent.CMD_G_GENERIC},
	}...)
}
