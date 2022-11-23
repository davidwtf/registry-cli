package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func delCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del IMAGE_REF",
		Short:   "delete the manifest",
		Example: `  registrycli del 127.0.0.1:5000/repo1:v1.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.ErrNeedImageReference
			}
			if len(args) > 1 {
				return errors.ErrTooManyArgs
			}

			if err := opts.ParseReference(args[0]); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Del(opts)
		},
	}
	cmd.Flags().BoolVar(&opts.Untag, "untag", false, "untag the tag")
	return cmd
}
