// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package query

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/flameql"
	"github.com/pyroscope-io/pyroscope/pkg/model"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"

	"github.com/erda-project/erda/pkg/http/httpserver"
)

type maxNodesKeyType int

const currentMaxNodes maxNodesKeyType = iota

type renderParams struct {
	format   string
	maxNodes int
	gi       *storage.GetInput

	leftStartTime time.Time
	leftEndTime   time.Time
	rghtStartTime time.Time
	rghtEndTime   time.Time
}

type renderMetadataResponse struct {
	flamebearer.FlamebearerMetadataV1
	AppName   string `json:"appName"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
	Query     string `json:"query"`
	MaxNodes  int    `json:"maxNodes"`
}

type annotationsResponse struct {
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

type renderResponse struct {
	flamebearer.FlamebearerProfile
	Metadata    renderMetadataResponse `json:"metadata"`
	Annotations []annotationsResponse  `json:"annotations"`
}

func (p *provider) render(rw http.ResponseWriter, r *http.Request) {
	var req renderParams
	if err := p.renderParametersFromRequest(r, &req); err != nil {
		httpserver.WriteErr(rw, "400", err.Error())
		return
	}
	out, err := p.st.Get(r.Context(), req.gi)
	if err != nil {
		httpserver.WriteErr(rw, "400", err.Error())
		return
	}
	if out == nil && p.isQueryWithProfileID(req.gi) {
		p.removeProfileIDMatcher(req.gi)
		out, err = p.st.Get(r.Context(), req.gi)
	}

	if out == nil {
		out = &storage.GetOutput{
			Tree:     tree.New(),
			Timeline: segment.GenerateTimeline(req.gi.StartTime, req.gi.EndTime),
		}
	}

	var appName string
	if req.gi.Key != nil {
		appName = req.gi.Key.AppName()
	} else if req.gi.Query != nil {
		appName = req.gi.Query.AppName
	}
	filename := fmt.Sprintf("%v %v", appName, req.gi.StartTime.UTC().Format(time.RFC3339))

	switch req.format {
	case "json":
		fallthrough
	default:
		flame := flamebearer.NewProfile(flamebearer.ProfileConfig{
			Name:      filename,
			MaxNodes:  req.maxNodes,
			Tree:      out.Tree,
			Timeline:  out.Timeline,
			Groups:    out.Groups,
			Telemetry: out.Telemetry,
			Metadata: metadata.Metadata{
				SpyName:         out.SpyName,
				SampleRate:      out.SampleRate,
				Units:           out.Units,
				AggregationType: out.AggregationType,
			},
		})

		res := p.mountRenderResponse(flame, appName, req.gi, req.maxNodes, []model.Annotation{})
		renderCounter.WithLabelValues(req.gi.Query.AppName).Inc()
		httpserver.WriteData(rw, res)
	}
}

func (p *provider) isQueryWithProfileID(gi *storage.GetInput) bool {
	if gi.Query != nil {
		for _, matcher := range gi.Query.Matchers {
			if matcher.Key == segment.ProfileIDLabelName {
				return true
			}
		}
	}
	return false
}

func (p *provider) removeProfileIDMatcher(gi *storage.GetInput) {
	matchers := make([]*flameql.TagMatcher, 0)
	for _, matcher := range gi.Query.Matchers {
		if matcher.Key != segment.ProfileIDLabelName {
			matchers = append(matchers, matcher)
		}
	}
	gi.Query.Matchers = matchers
}

func (p *provider) renderParametersFromRequest(r *http.Request, req *renderParams) error {
	v := r.URL.Query()
	req.gi = new(storage.GetInput)

	k := v.Get("name")
	q := v.Get("query")
	req.gi.GroupBy = v.Get("groupBy")

	switch {
	case k == "" && q == "":
		return fmt.Errorf("'query' or 'name' parameter is required")
	case k != "":
		sk, err := segment.ParseKey(k)
		if err != nil {
			return fmt.Errorf("name: parsing storage key: %v", err)
		}
		req.gi.Key = sk
	case q != "":
		qry, err := flameql.ParseQuery(q)
		if err != nil {
			return fmt.Errorf("query: %v", err)
		}
		req.gi.Query = qry
	}

	req.maxNodes = p.Cfg.MaxNodesRender
	if newMaxNodes, ok := MaxNodesFromContext(r.Context()); ok {
		req.maxNodes = newMaxNodes
	}
	if mn, err := strconv.Atoi(v.Get("max-nodes")); err == nil && mn != 0 {
		req.maxNodes = mn
	}
	if mn, err := strconv.Atoi(v.Get("maxNodes")); err == nil && mn != 0 {
		req.maxNodes = mn
	}

	req.gi.StartTime = attime.Parse(v.Get("from"))
	req.gi.EndTime = attime.Parse(v.Get("until"))
	req.format = v.Get("format")

	return expectFormats(req.format)
}

func (p *provider) mountRenderResponse(flame flamebearer.FlamebearerProfile, appName string, gi *storage.GetInput, maxNodes int, annotations []model.Annotation) renderResponse {
	md := renderMetadataResponse{
		FlamebearerMetadataV1: flame.Metadata,
		AppName:               appName,
		StartTime:             gi.StartTime.Unix(),
		EndTime:               gi.EndTime.Unix(),
		Query:                 gi.Query.String(),
		MaxNodes:              maxNodes,
	}

	annotationsResp := make([]annotationsResponse, len(annotations))
	for i, an := range annotations {
		annotationsResp[i] = annotationsResponse{
			Content:   an.Content,
			Timestamp: an.Timestamp.Unix(),
		}
	}

	return renderResponse{
		FlamebearerProfile: flame,
		Metadata:           md,
		Annotations:        annotationsResp,
	}
}

func expectFormats(format string) error {
	switch format {
	case "json", "pprof", "collapsed", "":
		return nil
	default:
		return errors.New(fmt.Sprintf("unknown format: %s", format))
	}
}

func MaxNodesFromContext(ctx context.Context) (int, bool) {
	v, ok := ctx.Value(currentMaxNodes).(int)
	return v, ok
}
