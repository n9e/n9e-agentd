package cmds

import . "github.com/n9e/n9e-agentd/pkg/agent"

var (
	cmds = []CmdOps{
		{CmdFactory: newDocsCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newCompletionCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newEnvCmd, GroupNum: CMD_G_DEUBG},
		{CmdFactory: newVersionCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newConfigCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newCheckCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newIntegrationCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newOptionsCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStartCmd, GroupNum: CMD_G_RESOURCE},
	}
)

func init() {
	RegisterCmd(cmds...)
}
