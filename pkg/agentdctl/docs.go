// Copyright 2018,2019 freewheel. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package agentdctl

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const docsDesc = `
Generate documentation files for Agentdctl.

This command can generate documentation for Agentdctl in the following formats:

- Markdown
- Man pages

It can also generate bash autocompletions.

	$ agentdctl create docs --type markdown --dir mydocs/
	$ agentdctl create docs --type man --dir mymans/
	$ agentdctl create docs --type bash --dir /etc/bash_completion.d/
	$ agentdctl create docs --type zsh --dir myzsh/
`

type docsCmd struct {
	dest          string
	docTypeString string
	topCmd        *cobra.Command
}

func newDocsCmd(env *EnvSettings) *cobra.Command {
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
			Title:   "agentdctl",
			Section: "1",
		}
		return doc.GenManTree(d.topCmd, header, d.dest)
	case "bash":
		return d.topCmd.GenBashCompletionFile(filepath.Join(d.dest, "agentdctl.bash"))
	case "zsh":
		return d.topCmd.GenZshCompletionFile(filepath.Join(d.dest, "agentdctl.zsh"))
	default:
		return fmt.Errorf("unknown doc type %q. Try 'markdown' or 'man'", d.docTypeString)
	}
}

const completionDesc = `
Generate autocompletions script for fks for the specified shell (bash).

This command can generate shell autocompletions. e.g.

	$ agentdctl completion

Can be sourced as such

	$ source <(agentdctl completion)
`

func newCompletionCmd(env *EnvSettings) *cobra.Command {
	var typ string

	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate autocompletions script for the specified shell [bash/zsh] (defualt:bash)",
		Long:  completionDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				typ = args[0]
			}
			switch typ {
			case "zsh":
				return env.TopCmd.GenZshCompletion(env.Out)
			case "bash", "":
				return env.TopCmd.GenBashCompletion(env.Out)
			default:
				return fmt.Errorf("unsupported %s", typ)
			}
		},
	}
	return cmd
}
