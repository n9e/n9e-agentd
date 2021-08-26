package cmds

import (
	. "github.com/n9e/n9e-agentd/pkg/agent"
	"github.com/n9e/n9e-agentd/pkg/config"
)

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
		{CmdFactory: newConfigCheckCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newHostnameCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStopCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStreamLogsCmd, GroupNum: CMD_G_GENERIC},
	}
	loggerName config.LoggerName = "CORE"
)

func init() {
	RegisterCmd(cmds...)
}
