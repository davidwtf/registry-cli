package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func layerCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "layer LAYER_REF",
		Short:   "Get layer's content",
		Example: `  registrycli layer 127.0.0.1:5000/repo1@sha256:275b2e73e3dc5cbf88c41ba15962045f0d36eeaf09dfe01f259ff2a12d3326af`,
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

			return action.Layer(opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Destination, "destination", "d", "./layers", "location to save layer")
	return cmd
}
