package option

import (
	"context"
	"fmt"
	"io"
	"time"
)

const (
	TextOutput = "text"
	JSONOutput = "json"
)

var (
	AllOutputFormats = []string{
		TextOutput,
		JSONOutput,
	}
)

type Options struct {
	Username    string
	Password    string
	Auth        string
	Server      string
	Repositiory string
	Output      string
	Destination string
	Debug       bool
	AllRepos    bool
	Insecure    bool
	PlainHTTP   bool
	Parellel    int
	StdErr      io.Writer
	StdOut      io.Writer
	Ctx         context.Context
	Userdata    interface{}
}

func (opts *Options) IsSupportOutput(supports ...string) bool {
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

func (opts *Options) WriteDebug(msg string, err error) {
	if !(opts != nil && opts.Debug && opts.StdErr != nil) {
		return
	}

	opts.StdErr.Write([]byte(fmt.Sprintf("%s %s err: %v\n", time.Now().Format(time.RFC3339Nano), msg, err)))
}
