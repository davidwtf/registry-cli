package action

import (
	"context"
	"fmt"
	"net/http"
	"registry-cli/pkg/client"
	"registry-cli/pkg/option"

	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	registryapiv2 "github.com/docker/distribution/registry/api/v2"
	"github.com/opencontainers/go-digest"
)

func Del(tagOrDigest string, opts *option.Options) error {
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

	dgst := digest.Digest(tagOrDigest)

	if opts.Untag {
		if dgst.Validate() == nil {
			opts.WriteDebug("need a tag not digest", nil)
			return fmt.Errorf(`need a tag not digest: "%s"`, tagOrDigest)
		}
		if err := untag(opts.Ctx, cli, repo, tagOrDigest); err != nil {
			opts.WriteDebug(fmt.Sprintf(`untag "%s"`, tagOrDigest), err)
			return err
		}
	} else {
		manifestService, err := repo.Manifests(opts.Ctx)
		if err != nil {
			opts.WriteDebug("init mainifest service", err)
			return err
		}
		if dgst.Validate() != nil {
			_, err = manifestService.Get(opts.Ctx, "", distribution.WithTag(tagOrDigest), registryclient.ReturnContentDigest(&dgst))
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch digest for "%s"`, tagOrDigest), err)
				return err
			}
		}

		if err := manifestService.Delete(opts.Ctx, dgst); err != nil {
			opts.WriteDebug(fmt.Sprintf(`delete digest "%s"`, dgst.String()), err)
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
