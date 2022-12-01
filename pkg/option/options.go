package option

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

const (
	TextOutput = "text"
	JSONOutput = "json"

	SortByTag     = "tag"
	SortBySize    = "size"
	SortByCreated = "created"

	DefualtRegistry   = "docker.io"
	DefaultV2Registry = "registry-1.docker.io"
)

var (
	AllOutputFormats = []string{
		TextOutput,
		JSONOutput,
	}

	AllSortMethods = []string{
		SortByTag,
		SortBySize,
		SortByCreated,
	}
)

type Options struct {
	Username    string
	Password    string
	Auth        string
	Server      string
	Repositiory string
	Tag         string
	Digest      digest.Digest
	Output      string
	Sort        string
	Destination string
	Debug       bool
	ShowType    bool
	ShowDigest  bool
	ShowSummary bool
	Insecure    bool
	PlainHTTP   bool
	Untag       bool
	StdErr      io.Writer
	StdOut      io.Writer
	Ctx         context.Context
}

func (opts *Options) ParseReference(ref string) error {
	named, err := reference.ParseDockerRef(ref)
	if err != nil {
		return fmt.Errorf(`parse image reference "%s" error: %v`, ref, err)
	}
	opts.Server = reference.Domain(named)
	opts.Repositiory = reference.Path(named)
	if namedTaged, ok := named.(reference.NamedTagged); ok {
		opts.Tag = namedTaged.Tag()
	}
	if canonical, ok := named.(reference.Canonical); ok {
		opts.Digest = canonical.Digest()
	}

	if opts.Server == DefualtRegistry {
		opts.Server = DefaultV2Registry
	}

	return nil
}

func (opts *Options) IsSupportedOutput(supports ...string) bool {
	if opts == nil {
		return false
	}
	if supports == nil {
		supports = AllOutputFormats
	}
	for _, o := range supports {
		if o == opts.Output {
			return true
		}
	}
	return false
}

func (opts *Options) IsSupportedSort(supports ...string) bool {
	if opts == nil {
		return false
	}
	if supports == nil {
		supports = AllSortMethods
	}
	for _, o := range supports {
		if o == opts.Sort {
			return true
		}
	}
	return false
}

func (opts *Options) WriteDebug(msg string, err error) {
	if !(opts != nil && opts.Debug && opts.StdErr != nil) {
		return
	}

	opts.StdErr.Write([]byte(fmt.Sprintf("%s %s err: %v\n", time.Now().Format(time.RFC3339Nano), msg, err)))
}
