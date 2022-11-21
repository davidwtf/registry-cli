package action

import (
	"encoding/json"
	"fmt"
	"registry-cli/pkg/client"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"registry-cli/pkg/output"
	"sync"
	"time"

	"github.com/distribution/distribution/manifest/manifestlist"
	"github.com/distribution/distribution/manifest/ocischema"
	"github.com/distribution/distribution/manifest/schema1"
	"github.com/distribution/distribution/manifest/schema2"
	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	maxParellel = 20
)

var ()

type tagInfo struct {
	Repoistory string     `json:"repository"`
	Tag        string     `json:"tag"`
	Platform   string     `json:"platform"`
	Size       *int64     `json:"size"`
	Created    *time.Time `json:"created"`
	Type       string     `json:"type"`
	Digest     string     `json:"digest"`
}

func (t tagInfo) Header() []string {
	return []string{RepoHeader, TagHeader, PlatformHeader, SizeHeader, CreatedHeader, TypeHeader, DigestHeader}
}

func (t *tagInfo) Output(opts *option.Options) error {
	switch opts.Output {
	case option.JSONOutput:
		return opts.Userdata.(*output.JSONArrayWriter).Write(t)
	case option.TextOutput, option.DefaultOutput:
		return opts.Userdata.(*output.TextWriter).Write(
			t.Repoistory, t.Tag, t.Platform,
			output.SizeToShow(t.Size), output.TimeToShow(t.Created),
			t.Type, t.Digest)
	}
	return errors.ErrUnknownOutput
}

func Tags(opts *option.Options) error {
	var err error
	switch opts.Output {
	case option.JSONOutput:
		opts.Userdata, err = output.NewJSONArrayWriter(opts.StdOut)
	case option.TextOutput, option.DefaultOutput:
		opts.Userdata, err = output.NewTextWriter(opts.StdOut, tagInfo{}.Header()...)
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

	if opts.AllRepos {
		registry, err := cli.NewRegistry()
		if err != nil {
			opts.WriteDebug("init registry service", err)
			return err
		}

		if err := cli.WalkAllRepos(opts.Ctx, registry, func(repo string) (stop bool, err error) {
			err = fetchRepoTags(opts, cli, repo)
			if err != nil {
				return true, err
			}
			return false, nil
		}); err != nil {
			opts.WriteDebug("walk through all repoistories", err)
			return err
		}
	} else {
		err := fetchRepoTags(opts, cli, opts.Repositiory)
		if err != nil {
			return err
		}
	}

	if opts.Output == option.JSONOutput {
		err = opts.Userdata.(*output.JSONArrayWriter).Finish()
	}

	return err
}

func fetchRepoTags(opts *option.Options, cli *client.Client, repoName string) error {
	repo, err := cli.NewRepository(repoName, client.PullAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return err
	}

	mainifestService, err := repo.Manifests(opts.Ctx)
	if err != nil {
		opts.WriteDebug("init manifest service", err)
		return err
	}
	tags, err := repo.Tags(opts.Ctx).All(opts.Ctx)
	if err != nil {
		return err
	}
	if opts.Parellel && len(tags) > 1 {
		numParellel := maxParellel
		if numParellel > len(tags) {
			numParellel = len(tags)
		}
		ch := make(chan string)
		stop := make(chan bool)
		wg := sync.WaitGroup{}
		wg.Add(len(tags))
		for i := 0; i < numParellel; i++ {
			go fetchWorker(opts, repo, mainifestService, repoName, &wg, ch, stop)
		}

		for _, tag := range tags {
			ch <- tag
		}
		wg.Wait()
		close(stop)
	} else {
		for _, tag := range tags {
			if err := fetchTagInfo(opts, repo, mainifestService, repoName, tag); err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch gat info for "%s:%s"`, repoName, tag), err)
				continue
			}
		}
	}
	return nil
}

func fetchWorker(opts *option.Options, repo distribution.Repository, mainifestService distribution.ManifestService, repoName string, wg *sync.WaitGroup, ch <-chan string, stop <-chan bool) {
	for {
		select {
		case tag := <-ch:
			if err := fetchTagInfo(opts, repo, mainifestService, repoName, tag); err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch gat info for "%s:%s"`, repoName, tag), err)
			}
			wg.Done()
		case <-stop:
			return
		}
	}
}

func fetchTagInfo(opts *option.Options, repo distribution.Repository, mainifestService distribution.ManifestService, repoName, tag string) error {
	var dgst digest.Digest
	man, err := mainifestService.Get(opts.Ctx, "", distribution.WithTag(tag), registryclient.ReturnContentDigest(&dgst))
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`get manifest for "%s"`, tag), err)
		return err
	}
	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		for _, ref := range realMan.Manifests {
			man, err := mainifestService.Get(opts.Ctx, ref.Digest, registryclient.ReturnContentDigest(&dgst))
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get manifest for list "%s"'s "%s"`, tag, ref.Digest), err)
				return err
			}
			info, err := getManifestInfo(opts, repo, man)
			if err != nil {
				return err
			}
			info.Repoistory = repoName
			info.Tag = tag
			info.Digest = dgst.String()
			if err := info.Output(opts); err != nil {
				return err
			}
		}
	default:
		info, err := getManifestInfo(opts, repo, man)
		if err != nil {
			return err
		}
		info.Repoistory = repoName
		info.Tag = tag
		info.Digest = dgst.String()
		if err := info.Output(opts); err != nil {
			return err
		}
	}
	return nil
}

func getManifestInfo(opts *option.Options, repo distribution.Repository, manifest distribution.Manifest) (*tagInfo, error) {
	switch realMan := manifest.(type) {
	case *schema1.SignedManifest:
		size := int64(0)
		for _, layer := range realMan.FSLayers {
			blob, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, layer.BlobSum)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get blob "%s"`, layer.BlobSum), err)
				continue
			}
			size += int64(len(blob))
		}
		return &tagInfo{
			Type:     realMan.MediaType,
			Platform: realMan.Architecture,
			Size:     &size,
		}, nil
	case *schema2.DeserializedManifest:
		var created *time.Time
		platform := ""
		if realMan.Config.MediaType == schema2.MediaTypeImageConfig || realMan.Config.MediaType == ocispec.MediaTypeImageConfig {
			config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, realMan.Config.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, realMan.Config.Digest), err)
				return nil, err
			}
			image := ocispec.Image{}
			if err := json.Unmarshal(config, &image); err != nil {
				opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, realMan.Config.Digest), err)
				return nil, err
			}
			created = image.Created
			platform = fmt.Sprintf("%s/%s", image.OS, image.Architecture)
		}

		size := int64(0)
		for _, layer := range realMan.Layers {
			size += layer.Size
		}
		mediaType := realMan.MediaType
		if mediaType == "" {
			mediaType = realMan.Config.MediaType
		}
		return &tagInfo{
			Type:     mediaType,
			Created:  created,
			Platform: platform,
			Size:     &size,
		}, nil
	case *ocischema.DeserializedManifest:
		var created *time.Time
		platform := ""
		if realMan.Config.MediaType == schema2.MediaTypeImageConfig || realMan.Config.MediaType == ocispec.MediaTypeImageConfig {
			config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, realMan.Config.Digest)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, realMan.Config.Digest), err)
				return nil, err
			}
			image := ocispec.Image{}
			if err := json.Unmarshal(config, &image); err != nil {
				opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, realMan.Config.Digest), err)
				return nil, err
			}
			created = image.Created
			platform = fmt.Sprintf("%s/%s", image.OS, image.Architecture)
		}

		size := int64(0)
		for _, layer := range realMan.Layers {
			size += layer.Size
		}
		mediaType := realMan.MediaType
		if mediaType == "" {
			mediaType = realMan.Config.MediaType
		}
		return &tagInfo{
			Type:     mediaType,
			Created:  created,
			Platform: platform,
			Size:     &size,
		}, nil
	}
	return nil, errors.ErrUnknownManifest
}
