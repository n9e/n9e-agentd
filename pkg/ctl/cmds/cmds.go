package cmds

import "github.com/n9e/n9e-agentd/pkg/ctl"

var (
	cmds = []ctl.CmdOps{
		{newDocsCmd, ctl.CMD_G_OTHER},
		{newCompletionCmd, ctl.CMD_G_OTHER},
		{newEnvCmd, ctl.CMD_G_OTHER},
		{newVersionCmd, ctl.CMD_G_OTHER},
		{newConfigCmd, ctl.CMD_G_OTHER},
	}
)

func init() {
	ctl.RegisterCmd(cmds...)
}
