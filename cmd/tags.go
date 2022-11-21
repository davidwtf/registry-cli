package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func tagsCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Short:   "list tags",
		Example: `  registrycli tags -s 127.0.0.1:5000 -r repo1`,
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
	cmd.Flags().BoolVar(&opts.AllRepos, "all", false, "show all repositories")
	cmd.Flags().IntVar(&opts.Parellel, "parellel", 20, "workers to fetch tags in parallel, set it to 0 to fetch tags serially")
	return cmd
}
