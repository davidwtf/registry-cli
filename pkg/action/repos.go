package action

import (
	"fmt"
	"registry-cli/pkg/client"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"registry-cli/pkg/output"
)

func Repos(opts *option.Options) error {
	var err error
	var w *output.JSONArrayWriter
	switch opts.Output {
	case option.JSONOutput:
		w, err = output.NewJSONArrayWriter(opts.StdOut)
	case option.TextOutput:
		_, err = fmt.Fprintln(opts.StdOut, "REPOSITORY")
	default:
		return errors.ErrUnknownOutput
	}
	if err != nil {
		opts.WriteDebug("init output", err)
		return err
	}

	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}
	registry, err := cli.NewRegistry()
	if err != nil {
		opts.WriteDebug("init registry service", err)
		return err
	}

	if err := cli.WalkAllRepos(opts.Ctx, registry, func(repo string) (stop bool, err error) {
		if opts.Output == option.JSONOutput {
			err = w.Write(repo)
		} else {
			_, err = fmt.Fprintln(opts.StdOut, repo)
		}
		if err != nil {
			return true, err
		}
		return false, nil
	}); err != nil {
		opts.WriteDebug("walk through all repoistories", err)
		return err
	}

	if opts.Output == option.JSONOutput {
		err = opts.Userdata.(*output.JSONArrayWriter).Finish()
	}

	return err
}
