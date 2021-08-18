package cmds

import "github.com/n9e/n9e-agentd/pkg/agent"

var (
	cmds = []agent.CmdOps{
		{CmdFactory: newDocsCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newCompletionCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newEnvCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newVersionCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newConfigCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newCheckCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newIntegrationCmd, GroupNum: agent.CMD_G_OTHER},
		{CmdFactory: newOptionsCmd, GroupNum: agent.CMD_G_OTHER},
	}
)

func init() {
	agent.RegisterCmd(cmds...)
}
