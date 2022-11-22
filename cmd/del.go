package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func delCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "del TAG_OR_DIGEST",
		Short:   "delete the manifest",
		Example: `  registrycli del v1.0 -s 127.0.0.1:5000 -r repo1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.ErrNeedTagOrManifest
			}

			if err := needRepo(opts); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Del(args[0], opts)
		},
	}
	addRepoOpt(cmd, opts)
	cmd.Flags().BoolVar(&opts.Untag, "untag", false, "untag the tag")
	return cmd
}
