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
		{CmdFactory: newOptionsCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStartCmd, GroupNum: CMD_G_RESOURCE},
		{CmdFactory: newConfigCheckCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newHostnameCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStopCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStreamLogsCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newDiagnoseCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStatsdCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newStatusCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newHealthCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newListchecksCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newReloadChecksCmd, GroupNum: CMD_G_GENERIC},
		{CmdFactory: newFlareCmd, GroupNum: CMD_G_GENERIC},
	}
	loggerName    config.LoggerName = "CORE"
	jmxLoggerName config.LoggerName = "JMXFETCH"
)

func init() {
	RegisterCmd(cmds...)
}
