package ctl

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

const (
	_ = iota
	PRI_PKG
)

const (
	CMD_G_OTHER = iota
	CMD_G_RESOURCE
	CMD_G_DEUBG
	CMD_G_INTEGRATION
	CMD_G_SIZE
)

var (
	hookOps       = []*HookOps{}
	groupOps      [CMD_G_SIZE][]*HookOps
	groupMessages = [CMD_G_SIZE]string{
		"Other Commands",
		"Resource Management Commands",
		"Troubleshooting and Debugging Commands",
		"Integration",
	}
)

type CmdOps struct {
	CmdFactory func(*EnvSettings) *cobra.Command
	GroupNum   int
}

func RegisterCmd(in ...CmdOps) error {
	for i := range in {
		v := &HookOps{
			CmdFactory: in[i].CmdFactory,
			GroupNum:   in[i].GroupNum,
		}

		hookOps = append(hookOps, v)
		groupOps[v.GroupNum] = append(groupOps[v.GroupNum], v)
	}
	return nil

}

type HookOps struct {
	CmdFactory func(*EnvSettings) *cobra.Command
	Owner      string
	GroupNum   int
	Priority   int
	Data       interface{}
}

func RegisterHooks(in []HookOps) error {
	for i, _ := range in {
		v := &in[i]
		hookOps = append(hookOps, v)
		groupOps[v.GroupNum] = append(groupOps[v.GroupNum], v)
	}
	return nil
}

func GetHookGroups(env *EnvSettings) ([]*cobra.Command, templates.CommandGroups) {
	groups := [CMD_G_SIZE]templates.CommandGroup{}

	for i := 0; i < CMD_G_SIZE; i++ {
		groups[i].Message = groupMessages[i]
	}

	for i, bucket := range groupOps {
		for _, ops := range bucket {
			groups[i].Commands = append(groups[i].Commands, ops.CmdFactory(env))
		}
	}

	return groups[0].Commands, templates.CommandGroups(groups[1:])
}
