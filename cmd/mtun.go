package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/mritd/mmh/core"
	"github.com/spf13/cobra"
)

var tunLeftAddr, tunRightAddr string
var tunReverse bool
var tunDynamic string

var mtun = &cobra.Command{
	Use:   "mtun SERVER -l LEFT_ADDR -r RIGHT_ADDR [OPTIONS]",
	Short: "Open a ssh tunnel",
	Example: `  mtun -l 127.0.0.1:8080 -r baidu.com:80 prod11
  mtun -D prod11
  mtun -D 8080 prod11
  mtun -D=0.0.0.0:1080 prod11`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("dynamic") {
			if tunLeftAddr != "" || tunRightAddr != "" || tunReverse {
				fmt.Println("option -D cannot be used with -l/-r/--reverse")
				_ = cmd.Help()
				return
			}

			server := ""
			switch len(args) {
			case 0:
				name, ok := core.SelectServer()
				if !ok {
					return
				}
				server = name
			case 1:
				server = args[0]
			case 2:
				// pflag 不会消耗 -D 后的值，兼容 `mtun -D 8080 SERVER` 写法
				tunDynamic = args[0]
				server = args[1]
			default:
				_ = cmd.Help()
				return
			}
			core.DynamicTunnel(server, tunDynamic)
			return
		}

		if len(args) > 1 || tunLeftAddr == "" || tunRightAddr == "" {
			_ = cmd.Help()
			return
		}
		server := ""
		if len(args) == 1 {
			server = args[0]
		} else {
			name, ok := core.SelectServer()
			if !ok {
				return
			}
			server = name
		}
		core.Tunnel(server, tunLeftAddr, tunRightAddr, tunReverse)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) (res []string, _ cobra.ShellCompDirective) {
		for _, s := range core.ListServers(true) {
			res = append(res, fmt.Sprintf("%s\tfrom %s(%s)", s.Name, filepath.Base(s.ConfigPath), s.Name))
		}
		return res, cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	mtun.PersistentFlags().StringVarP(&tunLeftAddr, "left", "l", "", "left address")
	mtun.PersistentFlags().StringVarP(&tunRightAddr, "right", "r", "", "right address")
	mtun.PersistentFlags().BoolVar(&tunReverse, "reverse", false, "reverse tcp tunnel(right to left)")
	mtun.PersistentFlags().StringVarP(&tunDynamic, "dynamic", "D", "", "dynamic port forwarding(socks5 proxy), default listen at "+core.DefaultSocksListenAddr)
	mtun.PersistentFlags().Lookup("dynamic").NoOptDefVal = core.DefaultSocksListenAddr
	_ = mtun.RegisterFlagCompletionFunc("left", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"127.0.0.1"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	_ = mtun.RegisterFlagCompletionFunc("right", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"127.0.0.1"}, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	rootCmd.AddCommand(mtun)
}
