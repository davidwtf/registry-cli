package client

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"registry-cli/pkg/errors"
	"registry-cli/pkg/option"
	"strings"

	"github.com/containers/image/v5/pkg/docker/config"
	"github.com/containers/image/v5/types"
	"github.com/distribution/distribution/registry/client/auth"
)

type credstore struct {
	username      string
	password      string
	auth          string
	refreshTokens map[string]string
}

func (cs *credstore) Basic(*url.URL) (string, string) {
	return cs.username, cs.password
}

func (cs *credstore) RefreshToken(u *url.URL, service string) string {
	return cs.refreshTokens[service]
}

func (cs *credstore) SetRefreshToken(u *url.URL, service string, token string) {
	if cs.refreshTokens != nil {
		cs.refreshTokens[service] = token
	}
}

func makeAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
}

func decodeAuth(auth string, opts *option.Options) (username, password string) {
	decoded, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		opts.WriteDebug("base64 decode auth", err)
		return "", ""
	}
	splited := strings.SplitN(string(decoded), ":", 2)
	if len(splited) == 2 {
		return splited[0], splited[1]
	}
	opts.WriteDebug("split username and password", errors.ErrSplitAuth)
	return "", ""
}

func NewCredStore(opts *option.Options) auth.CredentialStore {
	cs := credstore{
		refreshTokens: make(map[string]string),
	}
	if opts.Auth != "" {
		cs.username, cs.password = decodeAuth(opts.Auth, opts)
		cs.auth = opts.Auth
	} else if opts.Username != "" {
		cs.username, cs.password, cs.auth = opts.Username, opts.Password, makeAuth(opts.Username, opts.Password)
	} else {
		configs, err := config.GetAllCredentials(&types.SystemContext{})
		if err != nil {
			opts.WriteDebug("get system credentials", err)
		}
		if configs != nil {
			cfg, exist := configs[opts.Server]
			if exist {
				if cfg.IdentityToken != "" {
					cs.username, cs.password, cs.auth = cfg.Username, cfg.Password, cfg.IdentityToken
					if cs.username == "" {
						cs.username, cs.password = decodeAuth(cs.auth, opts)
					}
				} else if cfg.Username != "" {
					cs.username, cs.password, cs.auth = cfg.Username, cfg.Password, makeAuth(cfg.Username, cfg.Password)
				}
			}
		}
	}

	return &cs
}
