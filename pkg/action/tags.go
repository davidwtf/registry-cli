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

const maxWorkers = 100

type tagInfo struct {
	Tag      string     `json:"tag"`
	Platform string     `json:"platform"`
	Size     *int64     `json:"size"`
	Created  *time.Time `json:"created"`
	Type     string     `json:"type"`
	Digest   string     `json:"digest"`
}

func (t tagInfo) Header(opts *option.Options) []string {
	headers := []string{"TAG", "PLATFORM", "SIZE", "CREATED"}
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
		t.Tag, t.Platform,
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

type sum struct {
	Tags      int   `json:"tags"`
	Manfiests int   `json:"manifests"`
	Size      int64 `json:"size"`
}

type summary struct {
	Sum       sum             `json:"sum"`
	Platforms map[string]*sum `json:"platforms"`
}

type repoSummary struct {
	Repository string  `json:"repository"`
	Summary    summary `json:"summary"`
}

type repoInfo struct {
	repoSummary
	Tags []tagInfo `json:"tags"`
}

func Tags(opts *option.Options) error {
	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}

	n, tags, err := getTags(opts, cli)
	if err != nil {
		opts.WriteDebug("get tags", err)
		return err
	}

	if err := outputTags(opts, n, tags); err != nil {
		opts.WriteDebug("output tags", err)
		return err
	}

	return nil
}

func outputTags(opts *option.Options, num int, tags []tagInfo) error {
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

	repoInfo := repoInfo{
		repoSummary: repoSummary{
			Repository: opts.Repositiory,
			Summary: summary{
				Platforms: map[string]*sum{},
				Sum: sum{
					Tags: num,
				},
			},
		},
		Tags: tags,
	}

	for _, tag := range tags {
		size := int64(0)
		if tag.Size != nil {
			size = *tag.Size
		}
		repoInfo.Summary.Sum.Size += size
		repoInfo.Summary.Sum.Manfiests++
		if repoInfo.Summary.Platforms[tag.Platform] == nil {
			repoInfo.Summary.Platforms[tag.Platform] = &sum{}
		}
		repoInfo.Summary.Platforms[tag.Platform].Size += size
		repoInfo.Summary.Platforms[tag.Platform].Manfiests++
		repoInfo.Summary.Platforms[tag.Platform].Tags++
	}

	switch opts.Output {
	case option.JSONOutput:
		if err := output.WriteJSON(opts.StdOut, repoInfo); err != nil {
			return err
		}
	case option.TextOutput:
		w, err := output.NewTextWriter(opts.StdOut, tagInfo{}.Header(opts)...)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			if err := w.Write(tag.Column(opts)...); err != nil {
				return err
			}
		}
		if err := w.Flush(); err != nil {
			return err
		}

		if opts.ShowSummary {
			if _, err := fmt.Fprintln(opts.StdOut); err != nil {
				return err
			}
			if err := output.PrintStruct(opts.StdOut, repoInfo.repoSummary); err != nil {
				return err
			}
		}
	default:
		return errors.ErrUnknownOutput
	}
	return nil
}

func getTags(opts *option.Options, cli *client.Client) (int, []tagInfo, error) {
	repo, err := cli.NewRepository(opts.Repositiory, client.PullAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return 0, nil, err
	}

	manifestService, err := repo.Manifests(opts.Ctx)
	if err != nil {
		opts.WriteDebug("init manifest service", err)
		return 0, nil, err
	}

	tags, err := repo.Tags(opts.Ctx).All(opts.Ctx)
	if err != nil {
		opts.WriteDebug("get all tags", err)
		return 0, nil, err
	}

	numParellel := len(tags) / 5
	if numParellel < 1 {
		numParellel = 1
	}
	if numParellel > maxWorkers {
		numParellel = maxWorkers
	}
	inputCh := make(chan string)
	stop := make(chan bool)
	collectStopped := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(len(tags))

	var tagInfos []tagInfo
	resultCh := make(chan *tagInfo)
	for i := 0; i < numParellel; i++ {
		go fetchWorker(opts, repo, manifestService, &wg, inputCh, resultCh, stop)
	}
	go func() {
		for _, tag := range tags {
			inputCh <- tag
		}
	}()
	go collector(&tagInfos, resultCh, collectStopped)

	wg.Wait()
	close(stop)
	close(resultCh)
	<-collectStopped

	return len(tags), tagInfos, nil
}

func collector(result *[]tagInfo, resultCh <-chan *tagInfo, stop chan<- bool) {
	for {
		info := <-resultCh
		if info == nil {
			close(stop)
			break
		}
		*result = append(*result, *info)
	}
}

func fetchWorker(
	opts *option.Options,
	repo distribution.Repository,
	manifestService distribution.ManifestService,
	wg *sync.WaitGroup,
	inputCh <-chan string,
	resultCh chan<- *tagInfo,
	stop <-chan bool) {

	for {
		select {
		case tag := <-inputCh:
			infos, err := fetchTagInfos(opts, repo, manifestService, tag)
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`fetch get info for "%s"`, tag), err)
			} else {
				for _, info := range infos {
					resultCh <- info
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
	manifestService distribution.ManifestService,
	tag string) ([]*tagInfo, error) {

	var dgst digest.Digest
	man, err := manifestService.Get(opts.Ctx, "", distribution.WithTag(tag), registryclient.ReturnContentDigest(&dgst))
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`get manifest for "%s"`, tag), err)
		return nil, err
	}
	var r []*tagInfo
	switch realMan := man.(type) {
	case *manifestlist.DeserializedManifestList:
		for _, ref := range realMan.Manifests {
			man, err := manifestService.Get(opts.Ctx, ref.Digest, registryclient.ReturnContentDigest(&dgst))
			if err != nil {
				opts.WriteDebug(fmt.Sprintf(`get manifest for list "%s"'s "%s"`, tag, ref.Digest), err)
				return nil, err
			}
			info, err := getManifestInfo(opts, repo, man)
			if err != nil {
				return nil, err
			}
			info.Tag = tag
			info.Digest = dgst.String()
			r = append(r, info)
		}
	default:
		info, err := getManifestInfo(opts, repo, man)
		if err != nil {
			return nil, err
		}
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
			image, err := getImage(opts, repo, realMan.Config.Digest)
			if err != nil {
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
			image, err := getImage(opts, repo, realMan.Config.Digest)
			if err != nil {
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

func getImage(opts *option.Options, repo distribution.Repository, dgst digest.Digest) (*ocispec.Image, error) {
	config, err := repo.Blobs(opts.Ctx).Get(opts.Ctx, dgst)
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`get config for "%s"`, dgst), err)
		return nil, err
	}
	image := ocispec.Image{}
	if err := json.Unmarshal(config, &image); err != nil {
		opts.WriteDebug(fmt.Sprintf(`unmarshal config for "%s"`, dgst), err)
		return nil, err
	}
	return &image, nil
}
