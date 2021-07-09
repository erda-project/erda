package steve

import (
	"net/http"

	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/apiserver/pkg/urlbuilder"
)

func NewPrefixed(r *http.Request, schemas *types.APISchemas, prefix string) (types.URLBuilder, error) {
	requestURL := urlbuilder.ParseRequestURL(r)
	responseURLBase, err := urlbuilder.ParseResponseURLBase(requestURL, r)
	if err != nil {
		return nil, err
	}

	builder, err := urlbuilder.New(r, &urlbuilder.DefaultPathResolver{Prefix: prefix + "/v1"}, schemas)
	if err != nil {
		return nil, err
	}

	prefixedBuilder := &PrefixedURLBuilder{
		URLBuilder: builder,
		prefix:     prefix,
		base:       responseURLBase,
	}
	return prefixedBuilder, nil
}

type PrefixedURLBuilder struct {
	types.URLBuilder

	prefix string
	base   string
}

func (u *PrefixedURLBuilder) RelativeToRoot(path string) string {
	return urlbuilder.ConstructBasicURL(u.base, u.prefix, path)
}
