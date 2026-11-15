package cmd

import (
	"strings"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var copy2Group bool

var mcp = &cobra.Command{
	Use:   "mcp [-t] FILE/DIR|SERVER:PATH SERVER:PATH|FILE/DIR",
	Short: "Copies files between hosts on a network",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			_ = cmd.Help()
			return
		}
		resolved, ok := resolveCopyServer(args)
		if !ok {
			return
		}
		core.Copy(resolved, copy2Group)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, s.Name)
		}
		return res, cobra.ShellCompDirectiveDefault
	},
}

func init() {
	mcp.PersistentFlags().BoolVarP(&copy2Group, "tag", "t", false, "server tag")
	rootCmd.AddCommand(mcp)
}

// resolveCopyServer fills a missing server part interactively:
// "mcp :/remote/path ./local" downloads from the selected server,
// "mcp ./local /remote/path" uploads to the selected server.
func resolveCopyServer(args []string) ([]string, bool) {
	hasServer := false
	hasEmptyServer := false
	for _, a := range args {
		if strings.HasPrefix(a, ":") {
			hasEmptyServer = true
		} else if strings.Contains(a, ":") {
			hasServer = true
		}
	}
	if hasServer {
		return args, true
	}

	name, ok := core.SelectServer()
	if !ok {
		return nil, false
	}

	if hasEmptyServer {
		for i, a := range args {
			if strings.HasPrefix(a, ":") {
				args[i] = name + a
			}
		}
		return args, true
	}

	args[len(args)-1] = name + ":" + args[len(args)-1]
	return args, true
}
