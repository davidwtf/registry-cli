package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"registry-cli/pkg/option"

	"github.com/distribution/distribution/reference"
	registryclient "github.com/distribution/distribution/registry/client"
	"github.com/distribution/distribution/registry/client/auth"
	"github.com/distribution/distribution/registry/client/transport"
	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/client/auth/challenge"
)

const (
	bufsize = 50
)

type RepoHandler func(repoName string) (stop bool, err error)

type Client struct {
	challengeManager challenge.Manager
	credStore        auth.CredentialStore
	opts             *option.Options
	baseURL          string
	httpClient       *http.Client
}

func NewClient(opts *option.Options) (*Client, error) {
	var scheme string
	if opts.PlainHTTP {
		scheme = "http"
	} else {
		scheme = "https"
	}
	c := &Client{
		opts:             opts,
		challengeManager: challenge.NewSimpleManager(),
		credStore:        NewCredStore(opts),
		baseURL:          fmt.Sprintf("%s://%s", scheme, opts.Server),
		httpClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: opts.Insecure,
				},
			},
		},
	}

	if err := c.tryEstablishChallenges(); err != nil {
		opts.WriteDebug("failed to establish challegenes", err)
		return nil, err
	}

	return c, nil
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func (c *Client) NewRegistry() (registryclient.Registry, error) {
	roundTripper, err := c.GetRoundTripper("", CatalogAction)
	if err != nil {
		c.opts.WriteDebug("failed to get round tripper for catalog", err)
		return nil, err
	}
	return registryclient.NewRegistry(c.baseURL, roundTripper)
}

func (c *Client) NewRepository(repo string, action Action) (distribution.Repository, error) {
	repoNamed, err := reference.WithName(repo)
	if err != nil {
		c.opts.WriteDebug(fmt.Sprintf(`failed to refer name: "%s"`, repo), err)
		return nil, err
	}

	roundTripper, err := c.GetRoundTripper(repoNamed.String(), action)
	if err != nil {
		c.opts.WriteDebug(fmt.Sprintf(`failed to get round tripper for: "%s"`, repo), err)
		return nil, err
	}

	repository, err := registryclient.NewRepository(repoNamed, c.baseURL, roundTripper)
	if err != nil {
		c.opts.WriteDebug(fmt.Sprintf(`failed to create repository service for : "%s"`, repo), err)
		return nil, err
	}
	return repository, nil
}

func (c *Client) tryEstablishChallenges() error {
	endpointURL, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}

	endpointURL.Path = "/v2/"
	challenges, err := c.challengeManager.GetChallenges(*endpointURL)
	if err != nil {
		return err
	}

	if len(challenges) > 0 {
		return nil
	}

	resp, err := c.httpClient.Get(endpointURL.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := c.challengeManager.AddResponse(resp); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetRoundTripper(scope string, action Action) (http.RoundTripper, error) {
	return transport.NewTransport(c.httpClient.Transport,
		auth.NewAuthorizer(c.challengeManager,
			auth.NewBasicHandler(c.credStore),
			auth.NewTokenHandler(c.httpClient.Transport, c.credStore, scope, string(action)))), nil
}

func (c *Client) WalkAllRepos(ctx context.Context, registry registryclient.Registry, fun RepoHandler) error {
	buf := make([]string, bufsize)
	last := ""
	for {
		n, retErr := registry.Repositories(ctx, buf, last)
		if retErr != nil && retErr != io.EOF {
			return retErr
		}
		if n <= 0 {
			break
		}
		for _, repo := range buf[:n] {
			stop, err := fun(repo)
			if err != nil {
				return err
			}
			if stop {
				return nil
			}
		}
		if retErr == io.EOF {
			break
		}
		last = buf[n-1]
	}
	return nil
}
