package agentdctl

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func mkdir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return err
}

func newConfigZshCmd(env *EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "set zsh completion config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configZshCompletion(env)
		},
	}
}

const zshConfigUsage = `# Zsh completion file has been generated(~/.n9e/zsh/_n9e)

# So, to initialize the compsys add the following code into your ~/.zshrc file
cat >> ~/.zshrc <<'EOF'
# add custom completion scripts
fpath=(~/.n9e/zsh $fpath)

# compsys initialization
autoload -U compinit
compinit

# show completion menu when number of 
zstyle ':completion:*' menu select=2
EOF
`

func configZshCompletion(env *EnvSettings) error {
	if err := env.TopCmd.GenZshCompletion(env.Out); err != nil {
		return err
	}

	fmt.Fprintf(env.Out, zshConfigUsage)

	return nil
}

func newConfigBashCmd(env *EnvSettings) *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "set bash completion config",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configBashCompletion(env)
		},
	}
}

const bashConfigUsage = `# Bash completion file has been generated(~/.n9e/n9e.bash)

# So, to initialize the compsys add the following code into your ~/.bashrc file
cat >> ~/.bashrc <<'EOF'
if [ -f ~/.n9e/n9e.bash ]; then
	source ~/.n9e/n9e.bash
fi
EOF
`

func configBashCompletion(env *EnvSettings) error {
	if err := env.TopCmd.GenBashCompletion(env.Out); err != nil {
		return err
	}

	fmt.Fprintf(env.Out, bashConfigUsage)
	return nil
}
