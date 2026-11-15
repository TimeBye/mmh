package cmd

import (
	"fmt"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var mcx = &cobra.Command{
	Use:   "mcx",
	Short: "Change config file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			name, ok := core.SelectConfig()
			if !ok {
				return
			}
			core.SetConfig(name)
			fmt.Printf("👉 mmh changed config to [%s.yaml]...\n", name)
			return
		}

		core.SetConfig(args[0])
		fmt.Printf("👉 mmh changed config to [%s.yaml]...\n", args[0])
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) (res []string, _ cobra.ShellCompDirective) {
		for _, info := range core.Configs {
			res = append(res, fmt.Sprintf("%s\t%s", info.Name, info.Path))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(mcx)
}
