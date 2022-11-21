package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func inspectCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inspect [command options] TAG_OR_DIGEST",
		Short:   "inspect the tag or digest details",
		Example: `  inspect -s 127.0.0.1:5000 -r repo1 v1.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.ErrNeedTagOrManifest
			}

			if err := needRepo(opts); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Inspect(args[0], opts)
		},
	}
	addRepoOpt(cmd, opts)
	addOutputOpt(cmd, opts)

	return cmd
}
