package agentdctl

const (
	pkgName = "agentdctl"
)

var (
	_hookOps = GenHookOps2(pkgName, PRI_PKG, []HookOps2{
		{newDocsCmd, CMD_CREATE, 0},
		{newCompletionCmd, CMD_ROOT, CMD_G_SETTINGS},
		{newConfigBashCmd, CMD_CONFIG, 0},
		{newConfigZshCmd, CMD_CONFIG, 0},
		{newEnvCmd, CMD_ROOT, CMD_G_SETTINGS},
	})
)

func init() {
	for i := 0; i < CMD_SIZE; i++ {
		hookOps[i] = HookOpsBucket([]*HookOps{})
	}
	RegisterHooks(_hookOps)
}
