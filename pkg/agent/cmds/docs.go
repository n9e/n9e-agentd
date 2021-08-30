package cmds

import (
	"fmt"
	"path/filepath"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/yubo/apiserver/pkg/cmdcli"
)

const docsDesc = `
Generate documentation files for agent.

This command can generate documentation for agent in the following formats:

- Markdown
- Man pages

It can also generate bash autocompletions.

	$ agent create docs --type markdown --dir mydocs/
	$ agent create docs --type man --dir mymans/
	$ agent create docs --type bash --dir /etc/bash_completion.d/
	$ agent create docs --type zsh --dir myzsh/
`

type docsCmd struct {
	Dest          string `flag:"dir" default:"./" description:"directory to which documentation is written"`
	DocTypeString string `flag:"type" default:"markdown" description:"the type of documentation to generate (markdown, man, bash, zsh)"`
	topCmd        *cobra.Command
}

func newDocsCmd(env *agent.EnvSettings) *cobra.Command {
	dc := &docsCmd{}

	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation as markdown or man pages",
		Long:   docsDesc,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			dc.topCmd = env.TopCmd
			dc.topCmd.DisableAutoGenTag = true
			return dc.run()
		},
	}

	cmdcli.AddFlags(cmd.Flags(), dc)

	return cmd
}

func (d *docsCmd) run() error {
	switch d.DocTypeString {
	case "markdown", "mdown", "md":
		return doc.GenMarkdownTree(d.topCmd, d.Dest)
	case "man":
		header := &doc.GenManHeader{
			Title:   "agent",
			Section: "1",
		}
		return doc.GenManTree(d.topCmd, header, d.Dest)
	case "bash":
		return d.topCmd.GenBashCompletionFile(filepath.Join(d.Dest, "agent.bash"))
	case "zsh":
		return d.topCmd.GenZshCompletionFile(filepath.Join(d.Dest, "agent.zsh"))
	default:
		return fmt.Errorf("unknown doc type %q. Try 'markdown' or 'man'", d.DocTypeString)
	}
}
