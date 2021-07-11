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

package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/erda-project/erda-infra/pkg/transport/http/runtime"
)

// Spec .
type Spec struct {
	Path        string
	Method      string
	Service     string
	ServiceURL  string
	BackendPath string
	Attributes  map[string]interface{}
	Doc         string
	Handler     http.HandlerFunc
}

// Context .
type Context interface {
	Spec() *Spec
	Matcher() runtime.Matcher
	BackendMatcher() runtime.Matcher
	BackendPath(vars map[string]string) string
	ServiceURL() *url.URL
}

type apiContext struct {
	spec        *Spec
	backendURL  *url.URL
	serviceURL  *url.URL
	pubMatcher  runtime.Matcher
	backMatcher runtime.Matcher
	segments    []*pathSegment
}

// NewContext .
func NewContext(spec *Spec) (Context, error) {
	pubMatcher, err := runtime.Compile(spec.Path)
	if err != nil {
		return nil, fmt.Errorf("invalid path %q of service %q: %s", spec.Path, spec.Service, err)
	}
	var backMatcher runtime.Matcher
	var segments []*pathSegment
	if len(spec.BackendPath) > 0 {
		backMatcher, err = runtime.Compile(spec.BackendPath)
		if err != nil {
			return nil, fmt.Errorf("invalid BackendPath %q of service %q: %s", spec.BackendPath, spec.Service, err)
		}
		if !backMatcher.IsStatic() {
			segments = buildPathToSegments(backMatcher.Pattern())
		}
	}
	svrURL, err := url.Parse(spec.ServiceURL)
	if err != nil && spec.Handler == nil {
		return nil, fmt.Errorf("fail to parse service url: %s", err)
	}
	return &apiContext{
		spec:        spec,
		serviceURL:  svrURL,
		pubMatcher:  pubMatcher,
		backMatcher: backMatcher,
		segments:    segments,
	}, nil
}

func (c *apiContext) Spec() *Spec                     { return c.spec }
func (c *apiContext) ServiceURL() *url.URL            { return c.serviceURL }
func (c *apiContext) Matcher() runtime.Matcher        { return c.pubMatcher }
func (c *apiContext) BackendMatcher() runtime.Matcher { return c.backMatcher }
func (c *apiContext) BackendPath(vars map[string]string) string {
	if c.segments == nil {
		return c.spec.BackendPath
	}
	sb := strings.Builder{}
	for _, seg := range c.segments {
		if seg.typ == pathStatic {
			sb.WriteString(seg.name)
		} else {
			sb.WriteString(vars[seg.name])
		}
	}
	return sb.String()
}

type conetextKey struct{}

// WithContext .
func WithContext(parent context.Context, ctx Context) context.Context {
	return context.WithValue(parent, conetextKey{}, ctx)
}

// GetContext .
func GetContext(ctx context.Context) Context {
	v, _ := ctx.Value(conetextKey{}).(Context)
	return v
}
