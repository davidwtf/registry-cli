package option

import (
	"testing"
)

func TestParseReference(t *testing.T) {
	for _, c := range []struct {
		input  string
		server string
		repo   string
		tag    string
		digest string
		err    bool
	}{
		{
			input:  "alpine",
			server: "docker.io",
			repo:   "library/alpine",
			tag:    "latest",
		},
		{
			input:  "127.0.0.1:5000/repo1:v1",
			server: "127.0.0.1:5000",
			repo:   "repo1",
			tag:    "v1",
		},
		{
			input:  "127.0.0.1:5000/repo1/repo2@sha256:74f5f150164eb49b3e6f621751a353dbfbc1dd114eb9b651ef8b1b4f5cc0c0d5",
			server: "127.0.0.1:5000",
			repo:   "repo1/repo2",
			digest: "sha256:74f5f150164eb49b3e6f621751a353dbfbc1dd114eb9b651ef8b1b4f5cc0c0d5",
		},
	} {
		opts := &Options{}
		err := opts.ParseReference(c.input)
		if (err != nil) != c.err {
			if c.err {
				t.Error("expect err but get nill")
			} else {
				t.Errorf("unexceted error: %v", err)
			}
		} else {
			if opts.Server != c.server {
				t.Errorf("expect server %s, but got %s", c.server, opts.Server)
			}
			if opts.Repositiory != c.repo {
				t.Errorf("expect repository %s, but got %s", c.repo, opts.Repositiory)
			}
			if opts.Tag != c.tag {
				t.Errorf("expect tag %s, but got %s", c.tag, opts.Tag)
			}
			if opts.Digest.String() != c.digest {
				t.Errorf("expect digest %s, but got %s", c.digest, opts.Digest)
			}
		}
	}

}
