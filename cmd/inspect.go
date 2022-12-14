package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func inspectCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect IMAGE_REF",
		Short:   "inspect the manifest details",
		Example: `  registrycli inspect 127.0.0.1:5000/repo1:v1.0`,
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

			if err := opts.ParseReference(args[0]); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Inspect(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Output, "output", "o", option.TextOutput, "output format, options: json text")
	return cmd
}
