package main

import (
	"context"
	"net/url"
	"os"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"registry-cli/version"

	"github.com/spf13/cobra"
)

const (
	envServer = "REGISTRY_ADDRESS"
)

var subCmds = []func(*option.Options) *cobra.Command{
	reposCmd,
	tagsCmd,
	inspectCmd,
	delCmd,
	blobCmd,
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Server == "" {
				opts.Server = os.Getenv(envServer)
			}
			if opts.Server == "" {
				return errors.ErrNeedRegistry
			}
			if !checkHostPort(opts.Server) {
				return errors.ErrWrongRegistryFormat
			}
			return nil
		},
		Version: version.BuildVersion,
	}
	root.PersistentFlags().StringVarP(&opts.Server, "server", "s", "", "registry address, default use env REGISTRY_ADDRESS")
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

func addRepoOpt(cmd *cobra.Command, opts *option.Options) {
	cmd.Flags().StringVarP(&opts.Repositiory, "repository", "r", "", "repository")
}

func addOutputOpt(cmd *cobra.Command, opts *option.Options) {
	cmd.Flags().StringVarP(&opts.Output, "output", "o", option.TextOutput, "output format, options: json text")
}

func needRepo(opts *option.Options) error {
	if opts.Repositiory == "" {
		return errors.ErrNeedRepo
	}
	return nil
}

func checkHostPort(s string) bool {
	r, err := url.ParseRequestURI("https://" + s)
	if err != nil {
		return false
	}
	return r.Host == s
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
}
