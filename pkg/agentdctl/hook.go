package agentdctl

import (
	"sort"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"
)

// type {{{
type HookOps struct {
	Hook     func(*EnvSettings) *cobra.Command
	Owner    string
	HookNum  int
	GroupNum int
	Priority int
	Data     interface{}
}

type HookOps2 struct {
	Hook     func(*EnvSettings) *cobra.Command
	HookNum  int
	GroupNum int
}

type HookOpsBucket []*HookOps

func (p HookOpsBucket) Len() int {
	return len(p)
}

func (p HookOpsBucket) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p HookOpsBucket) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}

// }}}

const (
	_ = iota
	PRI_PKG
)

const (
	_ = iota
	CMD_ROOT
	CMD_ATTACH
	CMD_CHANGELOG
	CMD_CONFIG
	CMD_COPY
	CMD_CREATE
	CMD_DEBUG
	CMD_DELETE
	CMD_GET
	CMD_LOG
	CMD_METRICS
	CMD_MOVE
	CMD_PLAY
	CMD_RESET
	CMD_RESTART
	CMD_SHARE
	CMD_START
	CMD_STOP
	CMD_TEST
	CMD_UPDATE
	CMD_VERSION
	CMD_SIZE
)

const (
	CMD_G_OTHER = iota
	CMD_G_RESOURCE1
	CMD_G_RESOURCE2
	CMD_G_SETTINGS
	CMD_G_SERVICES
	CMD_G_SIZE
)

var (
	hookOps  [CMD_SIZE]HookOpsBucket
	groupOps [CMD_G_SIZE]HookOpsBucket
)

func RegisterHooks(in []HookOps) error {
	for i, _ := range in {
		v := &in[i]
		hookOps[v.HookNum] = append(hookOps[v.HookNum], v)
		if v.HookNum == CMD_ROOT {
			groupOps[v.GroupNum] = append(groupOps[v.GroupNum], v)
		}
	}
	return nil
}

func GetHookCmds(env *EnvSettings, cmd int) (ret []*cobra.Command) {
	bucket := hookOps[cmd]
	sort.Sort(bucket)
	for _, ops := range bucket {
		ret = append(ret, ops.Hook(env))
	}
	return
}

func GetHookGroups(env *EnvSettings) ([]*cobra.Command, templates.CommandGroups) {
	groups := [CMD_G_SIZE]templates.CommandGroup{
		{Message: "Other Commands:"},
		{Message: "Resource Management Commands (Beginner):"},
		{Message: "Resource Management Commands (Intermediate):"},
		{Message: "Settings Commands:"},
		{Message: "Independent service Commands:"},
	}

	for i, bucket := range groupOps {
		for _, ops := range bucket {
			groups[i].Commands = append(groups[i].Commands, ops.Hook(env))
		}
	}

	return groups[0].Commands, templates.CommandGroups(groups[1:])
}

func GenHookOps2(owner string, pri int, in []HookOps2) (out []HookOps) {
	for _, v := range in {
		out = append(out, HookOps{
			Hook:     v.Hook,
			HookNum:  v.HookNum,
			GroupNum: v.GroupNum,
			Owner:    owner,
			Priority: pri,
		})
	}
	return
}
