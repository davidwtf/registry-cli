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
)

const (
	HelmChartConfigMediaType = "application/vnd.cncf.helm.config.v1+json"
)

var AllSupportedConfigMediaTypes = []string{
	schema2.MediaTypeImageConfig,
	ocispec.MediaTypeImageConfig,
	HelmChartConfigMediaType,
}

func IsSupportedConfigMediaTypes(mediaType string) bool {
	for _, v := range AllSupportedConfigMediaTypes {
		if v == mediaType {
			return true
		}
	}
	return false
}

type manifestList struct {
	Digest                   digest.Digest                          `json:"digest"`
	DeserializedManifestList *manifestlist.DeserializedManifestList `json:"deserializedManifest"`
	Items                    []interface{}                          `json:"items,omitempty"`
}

func (m *manifestList) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return output.WriteJSON(opts.StdOut, m)
	case option.TextOutput:
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
	case option.TextOutput:
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
	case option.TextOutput:
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
	case option.TextOutput:
		return output.PrintStruct(opts.StdOut, m)
	}
	return errors.ErrUnknownOutput
}

func Inspect(opts *option.Options) error {
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

	if opts.Tag != "" {
		man, err = manifestService.Get(opts.Ctx, "", distribution.WithTag(opts.Tag), registryclient.ReturnContentDigest(&opts.Digest))
	} else {
		man, err = manifestService.Get(opts.Ctx, opts.Digest)
	}
	if err != nil {
		opts.WriteDebug("fetch manifest", err)
		return err
	}

	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		m := manifestList{
			Digest:                   opts.Digest,
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
		o, err := getManifestForOutput(opts, repo, manifestService, man, opts.Digest)
		if err != nil {
			opts.WriteDebug(fmt.Sprintf(`get manifest "%s" for output`, opts.Digest), err)
			return err
		}
		return o.Output(opts)
	}
}

func getManifestForOutput(opts *option.Options, repo distribution.Repository, manifestService distribution.ManifestService, man distribution.Manifest, dgst digest.Digest) (manifestOutput, error) {
	switch realMan := man.(type) {
	case *schema1.SignedManifest:
		return &manifestV1{
			Digest:         dgst,
			SignedManifest: realMan,
		}, nil
	case *schema2.DeserializedManifest:
		image, chart, err := parseConfig(opts, repo, realMan.Config.MediaType, realMan.Config.Digest)
		if err != nil {
			return nil, err
		}
		return &manifestV2{
			Digest:               dgst,
			DeserializedManifest: realMan,
			Image:                image,
			Chart:                chart,
		}, nil
	case *ocischema.DeserializedManifest:
		image, chart, err := parseConfig(opts, repo, realMan.Config.MediaType, realMan.Config.Digest)
		if err != nil {
			return nil, err
		}
		return &manifestOCI{
			Digest:               dgst,
			DeserializedManifest: realMan,
			Image:                image,
			Chart:                chart,
		}, nil
	}

	return nil, errors.ErrUnknownManifest
}

func parseConfig(opts *option.Options, repo distribution.Repository, mediaType string, dgst digest.Digest) (image *ocispec.Image, chart *chart.Metadata, err error) {
	if IsSupportedConfigMediaTypes(mediaType) {
		var config []byte
		config, err = repo.Blobs(opts.Ctx).Get(opts.Ctx, dgst)
		if err != nil {
			opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, dgst), err)
		} else {
			switch mediaType {
			case schema2.MediaTypeImageConfig, ocispec.MediaTypeImageConfig:
				err = json.Unmarshal(config, &image)
			case HelmChartConfigMediaType:
				err = json.Unmarshal(config, &chart)
			}
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, dgst), err)
			}
		}
	} else {
		opts.WriteDebug(fmt.Sprintf(`unsupported media type "%s"`, mediaType), nil)
	}
	return
}
