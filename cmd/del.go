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
		Short:   "untag the tag or delete the digest",
		Example: `  registrycli del sha256:74f5f150164eb49b3e6f621751a353dbfbc1dd114eb9b651ef8b1b4f5cc0c0d5 -s 127.0.0.1:5000 -r repo1`,
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
	return cmd
}
