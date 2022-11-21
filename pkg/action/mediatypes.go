package action

import (
	"github.com/distribution/distribution/manifest/schema2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"helm.sh/helm/v3/pkg/registry"
)

var AllSupportedConfigMediaTypes = []string{
	schema2.MediaTypeImageConfig,
	ocispec.MediaTypeImageConfig,
	registry.ConfigMediaType,
}

func IsSupportedConfigMediaTypes(mediaType string) bool {
	for _, v := range AllSupportedConfigMediaTypes {
		if v == mediaType {
			return true
		}
	}
	return false
}
