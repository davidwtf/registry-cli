package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func tagsCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags REPO_REF",
		Short:   "list tags",
		Example: `  registrycli tags 127.0.0.1:5000/repo1`,
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

			if !opts.IsSupportedSort() {
				return errors.ErrUnknownSort
			}

			if err := opts.ParseReference(args[0]); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Tags(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Output, "output", "o", option.TextOutput, "output format, options: json text")
	cmd.Flags().BoolVar(&opts.ShowType, "show-type", false, "show media type when output with text format")
	cmd.Flags().BoolVar(&opts.ShowDigest, "show-digest", false, "show digest when output with text format")
	cmd.Flags().BoolVar(&opts.ShowSummary, "show-summary", true, "show summary when output with text format")
	cmd.Flags().StringVar(&opts.Sort, "sort", "tag", "sort method, options: tag size created")
	return cmd
}
