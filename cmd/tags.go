package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func tagsCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags [command options]",
		Short:   "list repository's tags",
		Example: `  tags -s 127.0.0.1:5000 -r repo1 v1.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.ErrTooManyArgs
			}

			if opts.AllRepos {
				if opts.Repositiory != "" {
					return errors.ErrConflictAllRepo
				}
			} else {
				if err := needRepo(opts); err != nil {
					return err
				}
			}

			setDefaultOpts(opts, cmd)

			return action.Tags(opts)
		},
	}
	addRepoOpt(cmd, opts)
	addOutputOpt(cmd, opts)
	cmd.Flags().StringVar(&opts.Platform, "platform", "", "only show specified platform")
	cmd.Flags().StringVar(&opts.Platform, "type", "", "only show specified media type")
	cmd.Flags().BoolVar(&opts.AllRepos, "all", false, "show all repositories")
	cmd.Flags().BoolVar(&opts.Parellel, "parellel", true, "fetch tags in parallel")
	return cmd
}
