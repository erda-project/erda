// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
