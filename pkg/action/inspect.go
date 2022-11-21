package action

import (
	"encoding/json"
	"fmt"
	"registry-cli/pkg/client"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"registry-cli/pkg/output"

	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/distribution/distribution/manifest/ocischema"
	"github.com/distribution/distribution/manifest/schema1"
	"github.com/distribution/distribution/manifest/schema2"
	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/registry"
)

type manifestList struct {
	Digest                   digest.Digest                          `json:"digest"`
	DeserializedManifestList *manifestlist.DeserializedManifestList `json:"deserializedManifest"`
	Items                    []interface{}                          `json:"items,omitempty"`
}

func (m *manifestList) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return output.WriteJSON(opts.StdOut, m)
	case option.TextOutput, option.DefaultOutput:
		return output.PrintStruct(opts.StdOut, m)
	}
	return errors.ErrUnknownOutput
}

type manifestOutput interface {
	Output(opts *option.Options) error
}

type manifestV1 struct {
	Digest         digest.Digest           `json:"digest"`
	SignedManifest *schema1.SignedManifest `json:"signedManifest"`
}

func (m *manifestV1) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return output.WriteJSON(opts.StdOut, m)
	case option.TextOutput, option.DefaultOutput:
		return output.PrintStruct(opts.StdOut, m)
	}
	return errors.ErrUnknownOutput
}

type manifestV2 struct {
	Digest               digest.Digest                 `json:"digest"`
	DeserializedManifest *schema2.DeserializedManifest `json:"deserializedManifest"`
	Image                *ocispec.Image                `json:"image,omitempty"`
	Chart                *chart.Metadata               `json:"chart,omitempty"`
}

func (m *manifestV2) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return output.WriteJSON(opts.StdOut, m)
	case option.TextOutput, option.DefaultOutput:
		return output.PrintStruct(opts.StdOut, m)
	}
	return errors.ErrUnknownOutput
}

type manifestOCI struct {
	Digest               digest.Digest                   `json:"digest"`
	DeserializedManifest *ocischema.DeserializedManifest `json:"deserializedManifest"`
	Image                *ocispec.Image                  `json:"image,omitempty"`
	Chart                *chart.Metadata                 `json:"chart,omitempty"`
}

func (m *manifestOCI) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return output.WriteJSON(opts.StdOut, m)
	case option.TextOutput, option.DefaultOutput:
		return output.PrintStruct(opts.StdOut, m)
	}
	return errors.ErrUnknownOutput
}

func Inspect(tagOrDigest string, opts *option.Options) error {
	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}
	repo, err := cli.NewRepository(opts.Repositiory, client.PullAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return err
	}

	manifestService, err := repo.Manifests(opts.Ctx)
	if err != nil {
		opts.WriteDebug("init manifest service", err)
		return err
	}

	var man distribution.Manifest

	dgst := digest.Digest(tagOrDigest)
	if dgst.Validate() == nil {
		man, err = manifestService.Get(opts.Ctx, digest.Digest(tagOrDigest), registryclient.ReturnContentDigest(&dgst))
		if err != nil {
			return err
		}
	} else {
		dgst = ""
		man, err = manifestService.Get(opts.Ctx, "", distribution.WithTag(tagOrDigest), registryclient.ReturnContentDigest(&dgst))
		if err != nil {
			return err
		}
	}

	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		m := manifestList{
			Digest:                   dgst,
			DeserializedManifestList: realMan,
		}
		for _, ref := range realMan.Manifests {
			man, err := manifestService.Get(opts.Ctx, ref.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch manifest "%s"`, ref.Digest), err)
				continue
			}
			o, err := getManifestForOutput(opts, repo, manifestService, man, ref.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get manifest "%s" for output`, ref.Digest), err)
				continue
			}
			m.Items = append(m.Items, o)
		}
		return m.Output(opts)
	default:
		o, err := getManifestForOutput(opts, repo, manifestService, man, dgst)
		if err != nil {
			opts.WriteDebug(fmt.Sprintf(`get manifest "%s" for output`, dgst), err)
			return err
		}
		return o.Output(opts)
	}
}

func getManifestForOutput(opts *option.Options, repo distribution.Repository, manifestService distribution.ManifestService, man distribution.Manifest, dgst digest.Digest) (manifestOutput, error) {
	switch realMan := man.(type) {
	case *schema1.SignedManifest:
		m := &manifestV1{
			Digest:         dgst,
			SignedManifest: realMan,
		}
		return m, nil
	case *schema2.DeserializedManifest:
		m := &manifestV2{
			Digest:               dgst,
			DeserializedManifest: realMan,
		}
		if IsSupportedConfigMediaTypes(realMan.Config.MediaType) {
			config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, realMan.Config.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, realMan.Config.Digest), err)
			} else {
				var err error
				switch realMan.Config.MediaType {
				case schema2.MediaTypeImageConfig, ocispec.MediaTypeImageConfig:
					err = json.Unmarshal(config, &m.Image)
				case registry.ConfigMediaType:
					err = json.Unmarshal(config, &m.Chart)
				}
				if err != nil {
					opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, realMan.Config.Digest), err)
				}
			}
		} else {
			opts.WriteDebug(fmt.Sprintf(`unsupported media type "%s"`, realMan.Config.MediaType), nil)
		}

		return m, nil
	case *ocischema.DeserializedManifest:
		m := manifestOCI{
			Digest:               dgst,
			DeserializedManifest: realMan,
		}
		if IsSupportedConfigMediaTypes(realMan.Config.MediaType) {
			config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, realMan.Config.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, realMan.Config.Digest), err)
			} else {
				var err error
				switch realMan.Config.MediaType {
				case schema2.MediaTypeImageConfig, ocispec.MediaTypeImageConfig:
					err = json.Unmarshal(config, &m.Image)
				case registry.ConfigMediaType:
					err = json.Unmarshal(config, &m.Chart)
				}
				if err != nil {
					opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, realMan.Config.Digest), err)
				}
			}
		} else {
			opts.WriteDebug(fmt.Sprintf(`unsupported media type "%s"`, realMan.Config.MediaType), nil)
		}

		return &m, nil

	}

	return nil, errors.ErrUnknownManifest
}
