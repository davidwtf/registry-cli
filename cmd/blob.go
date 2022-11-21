package main

import (
	"registry-cli/pkg/action"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	"github.com/spf13/cobra"
)

func blobCmd(opts *option.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "blob BLOB_ID",
		Short:   "Get the blob content",
		Example: `  registrycli blob sha256:275b2e73e3dc5cbf88c41ba15962045f0d36eeaf09dfe01f259ff2a12d3326af -s 127.0.0.1:5000 -r repo1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.ErrNeedBlobId
			}

			if err := needRepo(opts); err != nil {
				return err
			}

			setDefaultOpts(opts, cmd)

			return action.Blob(args[0], opts)
		},
	}
	addRepoOpt(cmd, opts)
	cmd.Flags().StringVarP(&opts.Destination, "destination", "d", "./blobs", "location to save blob")
	return cmd
}
