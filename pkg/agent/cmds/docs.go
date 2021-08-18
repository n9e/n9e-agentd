package cmds

import (
	"fmt"
	"path/filepath"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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
	dest          string
	docTypeString string
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

	f := cmd.Flags()
	f.StringVar(&dc.dest, "dir", "./", "directory to which documentation is written")
	f.StringVar(&dc.docTypeString, "type", "markdown", "the type of documentation to generate (markdown, man, bash, zsh)")

	return cmd
}

func (d *docsCmd) run() error {
	switch d.docTypeString {
	case "markdown", "mdown", "md":
		return doc.GenMarkdownTree(d.topCmd, d.dest)
	case "man":
		header := &doc.GenManHeader{
			Title:   "agent",
			Section: "1",
		}
		return doc.GenManTree(d.topCmd, header, d.dest)
	case "bash":
		return d.topCmd.GenBashCompletionFile(filepath.Join(d.dest, "agent.bash"))
	case "zsh":
		return d.topCmd.GenZshCompletionFile(filepath.Join(d.dest, "agent.zsh"))
	default:
		return fmt.Errorf("unknown doc type %q. Try 'markdown' or 'man'", d.docTypeString)
	}
}
