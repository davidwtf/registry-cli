package action

import (
	"context"
	"fmt"
	"net/http"
	"registry-cli/pkg/client"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"

	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	registryapiv2 "github.com/docker/distribution/registry/api/v2"
)

func Del(opts *option.Options) error {
	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}

	repo, err := cli.NewRepository(opts.Repositiory, client.DeleteAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return err
	}
	if opts.Untag {
		if opts.Tag == "" {
			opts.WriteDebug("need a tag", nil)
			return errors.ErrNeedTag
		}
		if err := untag(opts.Ctx, cli, repo, opts.Tag); err != nil {
			opts.WriteDebug(fmt.Sprintf(`untag "%s"`, opts.Tag), err)
			return err
		}
	} else {
		manifestService, err := repo.Manifests(opts.Ctx)
		if err != nil {
			opts.WriteDebug("init mainifest service", err)
			return err
		}
		if opts.Tag != "" {
			_, err = manifestService.Get(opts.Ctx, "", distribution.WithTag(opts.Tag), registryclient.ReturnContentDigest(&opts.Digest))
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch digest for "%s"`, opts.Tag), err)
				return err
			}
		}

		if err := manifestService.Delete(opts.Ctx, opts.Digest); err != nil {
			opts.WriteDebug(fmt.Sprintf(`delete digest "%s"`, opts.Digest), err)
			return err
		}
	}
	return nil
}

func untag(ctx context.Context, cli *client.Client, repo distribution.Repository, tag string) error {
	ref, err := reference.WithTag(repo.Named(), tag)
	if err != nil {
		return err
	}
	ub, err := registryapiv2.NewURLBuilderFromString(cli.GetBaseURL(), false)
	if err != nil {
		return err
	}
	u, err := ub.BuildManifestURL(ref)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}

	roundTriper, err := cli.GetRoundTripper(repo.Named().String(), client.DeleteAction)
	if err != nil {
		return err
	}

	resp, err := roundTriper.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if registryclient.SuccessStatus(resp.StatusCode) {
		return nil
	}
	return registryclient.HandleErrorResponse(resp)
}
