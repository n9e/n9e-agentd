package cmds

import (
	"fmt"
	"sort"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/apiserver/response"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func newTaggerListCmd(env *agent.EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "tagger",
		Short: "Print the tagger content of a running agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return taggerList(env)
		},
	}
}

func taggerList(env *agent.EnvSettings) error {
	resp := &response.TaggerListResponse{}

	if err := env.ApiCall("GET", "/api/v1/tagger", nil, nil, resp); err != nil {
		return err
	}

	for entity, tagItem := range resp.Entities {
		fmt.Fprintln(color.Output, fmt.Sprintf("\n=== Entity %s ===", color.GreenString(entity)))

		sources := make([]string, 0, len(tagItem.Tags))
		for source := range tagItem.Tags {
			sources = append(sources, source)
		}

		// sort sources for deterministic output
		sort.Slice(sources, func(i, j int) bool {
			return sources[i] < sources[j]
		})

		for _, source := range sources {
			fmt.Fprintln(color.Output, fmt.Sprintf("== Source %s ==", source))

			fmt.Fprint(color.Output, "Tags: [")

			// sort tags for easy comparison
			tags := tagItem.Tags[source]
			sort.Slice(tags, func(i, j int) bool {
				return tags[i] < tags[j]
			})

			for i, tag := range tags {
				tagInfo := strings.Split(tag, ":")
				fmt.Fprintf(color.Output, fmt.Sprintf("%s:%s", color.BlueString(tagInfo[0]), color.CyanString(strings.Join(tagInfo[1:], ":"))))
				if i != len(tags)-1 {
					fmt.Fprintf(color.Output, " ")
				}
			}

			fmt.Fprintln(color.Output, "]")
		}

		fmt.Fprintln(color.Output, "===")
	}

	return nil
}
