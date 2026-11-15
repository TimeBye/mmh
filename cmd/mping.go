package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var mping = &cobra.Command{
	Use:   "mping SERVER",
	Short: "Ping server",
	Run: func(cmd *cobra.Command, args []string) {
		switch len(args) {
		case 0:
			if name, ok := core.SelectServer(); ok {
				core.Ping(name)
			}
		case 1:
			core.Ping(args[0])
		default:
			_ = cmd.Help()
		}
	},
	ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, fmt.Sprintf("%s\tfrom %s(%s)", s.Name, filepath.Base(s.ConfigPath), s.Name))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	rootCmd.AddCommand(mping)
}
