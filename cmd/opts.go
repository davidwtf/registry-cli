package main

import (
	"context"
	"registry-cli/pkg/option"
	"registry-cli/version"

	"github.com/spf13/cobra"
)

var subCmds = []func(*option.Options) *cobra.Command{
	reposCmd,
	tagsCmd,
	inspectCmd,
	delCmd,
	layerCmd,
}

func rootCmd() *cobra.Command {
	opts := option.Options{}
	root := &cobra.Command{
		Use:              "registrycli",
		Long:             "",
		SilenceUsage:     true,
		SilenceErrors:    true,
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Version: version.BuildVersion,
	}
	root.PersistentFlags().StringVarP(&opts.Username, "username", "u", "", "registry username")
	root.PersistentFlags().StringVarP(&opts.Password, "password", "p", "", "registry password")
	root.PersistentFlags().StringVar(&opts.Auth, "auth", "", "registry auth, base64 encoded username:password")
	root.PersistentFlags().BoolVar(&opts.Insecure, "insecure", false, "use insecure tls")
	root.PersistentFlags().BoolVar(&opts.PlainHTTP, "plain-http", false, "use http without tls")

	root.PersistentFlags().BoolVar(&opts.Debug, "debug", false, "enable debug output")

	for _, c := range subCmds {
		root.AddCommand(c(&opts))
	}

	return root
}

func setDefaultOpts(opts *option.Options, cmd *cobra.Command) {
	if opts.StdErr == nil {
		opts.StdErr = cmd.ErrOrStderr()
	}
	if opts.StdOut == nil {
		opts.StdOut = cmd.OutOrStdout()
	}
	if opts.Ctx == nil {
		opts.Ctx = context.Background()
	}
	opts.Parellel = 50
}
