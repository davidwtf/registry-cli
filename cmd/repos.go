package main

import (
	"net/url"
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func reposCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repos REGISTRY_ADDRESS",
		Short:   "list all repository names",
		Example: `  registrycli repos 127.0.0.1:5000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.ErrNeedImageReference
			}
			if len(args) > 1 {
				return errors.ErrTooManyArgs
			}

			if !opts.IsSupportedOutput() {
				return errors.ErrUnknownOutput
			}

			if !checkServer(args[0]) {
				return errors.ErrWrongRegistryAddress
			}

			opts.Server = args[0]

			setDefaultOpts(opts, cmd)

			return action.Repos(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Output, "output", "o", option.TextOutput, "output format, options: json text")
	return cmd
}

func checkServer(s string) bool {
	r, err := url.ParseRequestURI("https://" + s)
	if err != nil {
		return false
	}
	return r.Host == s
}
