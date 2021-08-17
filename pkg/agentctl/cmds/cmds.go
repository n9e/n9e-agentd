package cmds

import "github.com/n9e/n9e-agentd/pkg/agentctl"

var (
	cmds = []agentctl.CmdOps{
		{CmdFactory: newDocsCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newCompletionCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newEnvCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newVersionCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newConfigCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newCheckCmd, GroupNum: agentctl.CMD_G_OTHER},
		{CmdFactory: newIntegrationCmd, GroupNum: agentctl.CMD_G_OTHER},
	}
)

func init() {
	agentctl.RegisterCmd(cmds...)
}
