package action

import (
	"encoding/json"
	"fmt"
	"registry-cli/pkg/client"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"registry-cli/pkg/output"
	"sort"
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

type tagInfo struct {
	Repoistory string     `json:"repository"`
	Tag        string     `json:"tag"`
	Platform   string     `json:"platform"`
	Size       *int64     `json:"size"`
	Created    *time.Time `json:"created"`
	Type       string     `json:"type"`
	Digest     string     `json:"digest"`
}

func (t tagInfo) Header(opts *option.Options) []string {
	headers := []string{"REPOSITORY", "TAG", "PLATFORM", "SIZE", "CREATED"}
	if opts.ShowType {
		headers = append(headers, "TYPE")
	}
	if opts.ShowDigest {
		headers = append(headers, "DIGEST")
	}
	return headers
}

func (t *tagInfo) Column(opts *option.Options) []string {
	col := []string{
		t.Repoistory, t.Tag, t.Platform,
		output.SizeToShow(t.Size), output.TimeToShow(t.Created),
	}
	if opts.ShowType {
		col = append(col, t.Type)
	}
	if opts.ShowDigest {
		col = append(col, t.Digest)
	}
	return col
}

func Tags(opts *option.Options) error {
	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}

	tags, err := getTags(opts, cli)
	if err != nil {
		opts.WriteDebug("get tags", err)
		return err
	}

	if err := outputTags(opts, tags); err != nil {
		opts.WriteDebug("output tags", err)
		return err
	}

	return nil
}

func outputTags(opts *option.Options, tags []tagInfo) error {
	switch opts.Sort {
	case option.SortByTag:
		sort.SliceStable(tags, func(i, j int) bool {
			return tags[i].Tag < tags[j].Tag
		})
	case option.SortBySize:
		sort.SliceStable(tags, func(i, j int) bool {
			if tags[i].Size == nil {
				return true
			}
			if tags[j].Size == nil {
				return false
			}
			return (*tags[i].Size) < (*tags[j].Size)
		})
	case option.SortByCreated:
		sort.SliceStable(tags, func(i, j int) bool {
			if tags[i].Created == nil {
				return true
			}
			if tags[j].Created == nil {
				return false
			}
			return (*tags[i].Created).Before(*tags[j].Created)
		})
	default:
		return errors.ErrUnknownSort
	}

	switch opts.Output {
	case option.JSONOutput:
		w, err := output.NewJSONArrayWriter(opts.StdOut)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			if err := w.Write(tag); err != nil {
				return err
			}
		}
		return w.Finish()
	case option.TextOutput:
		w, err := output.NewTextWriter(opts.StdOut, tagInfo{}.Header(opts)...)
		if err != nil {
			return err
		}

		sum := struct {
			TotalNumber int
			TotalSize   int64
		}{}

		for _, tag := range tags {
			if tag.Size != nil {
				sum.TotalSize += *tag.Size
			}
			sum.TotalNumber++
			if err := w.Write(tag.Column(opts)...); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}
		if opts.ShowSum {
			if _, err := fmt.Fprintln(opts.StdOut); err != nil {
				return err
			}
			if err := output.PrintStruct(opts.StdOut, sum); err != nil {
				return err
			}
		}
	default:
		return errors.ErrUnknownOutput
	}
	return nil
}

func getTags(opts *option.Options, cli *client.Client) ([]tagInfo, error) {
	repo, err := cli.NewRepository(opts.Repositiory, client.PullAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return nil, err
	}

	mainifestService, err := repo.Manifests(opts.Ctx)
	if err != nil {
		opts.WriteDebug("init manifest service", err)
		return nil, err
	}

	tags, err := repo.Tags(opts.Ctx).All(opts.Ctx)
	if err != nil {
		opts.WriteDebug("get all tags", err)
		return nil, err
	}

	numParellel := opts.Parellel
	if numParellel > len(tags) {
		numParellel = len(tags)
	}
	inputCh := make(chan string)
	stop := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(len(tags))

	var tagInfos []tagInfo
	resultCh := make(chan tagInfo)
	for i := 0; i < numParellel; i++ {
		go fetchWorker(opts, repo, mainifestService, &wg, inputCh, resultCh, stop)
	}
	go func() {
		for _, tag := range tags {
			inputCh <- tag
		}
	}()
	go collector(&tagInfos, resultCh, stop)

	wg.Wait()
	close(stop)

	return tagInfos, nil
}

func collector(result *[]tagInfo, resultCh <-chan tagInfo, stop <-chan bool) {
	for {
		select {
		case info := <-resultCh:
			*result = append(*result, info)
		case <-stop:
			return
		}
	}
}

func fetchWorker(
	opts *option.Options,
	repo distribution.Repository,
	mainifestService distribution.ManifestService,
	wg *sync.WaitGroup,
	inputCh <-chan string,
	resultCh chan<- tagInfo,
	stop <-chan bool) {

	for {
		select {
		case tag := <-inputCh:
			infos, err := fetchTagInfos(opts, repo, mainifestService, tag)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch get info for "%s"`, tag), err)
			} else {
				for _, info := range infos {
					resultCh <- *info
				}
			}
			wg.Done()
		case <-stop:
			return
		}
	}
}

func fetchTagInfos(
	opts *option.Options,
	repo distribution.Repository,
	mainifestService distribution.ManifestService,
	tag string) ([]*tagInfo, error) {

	var dgst digest.Digest
	man, err := mainifestService.Get(opts.Ctx, "", distribution.WithTag(tag), registryclient.ReturnContentDigest(&dgst))
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`get manifest for "%s"`, tag), err)
		return nil, err
	}
	var r []*tagInfo
	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		for _, ref := range realMan.Manifests {
			man, err := mainifestService.Get(opts.Ctx, ref.Digest, registryclient.ReturnContentDigest(&dgst))
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get manifest for list "%s"'s "%s"`, tag, ref.Digest), err)
				return nil, err
			}
			info, err := getManifestInfo(opts, repo, man)
			if err != nil {
				return nil, err
			}
			info.Repoistory = opts.Repositiory
			info.Tag = tag
			info.Digest = dgst.String()
			r = append(r, info)
		}
	default:
		info, err := getManifestInfo(opts, repo, man)
		if err != nil {
			return nil, err
		}
		info.Repoistory = opts.Repositiory
		info.Tag = tag
		info.Digest = dgst.String()
		r = append(r, info)
	}
	return r, nil
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
