package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func reposCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "repos",
		Short:   "list all repository names",
		Example: `  registrycli repos -s 127.0.0.1:5000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.ErrTooManyArgs
			}

			if !opts.IsSupportOutput() {
				return errors.ErrUnknownOutput
			}

			setDefaultOpts(opts, cmd)

			return action.Repos(opts)
		},
	}
	addOutputOpt(cmd, opts)

	return cmd
}
