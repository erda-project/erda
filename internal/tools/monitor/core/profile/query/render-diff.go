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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pyroscope-io/pyroscope/pkg/flameql"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/metadata"
	"github.com/pyroscope-io/pyroscope/pkg/storage/tree"
	"github.com/pyroscope-io/pyroscope/pkg/structs/flamebearer"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"

	"github.com/erda-project/erda/pkg/http/httpserver"
)

type diffParams struct {
	Left  storage.GetInput
	Right storage.GetInput

	Format   string
	MaxNodes int
}

// RenderDiffResponse refers to the response of the renderDiff
type RenderDiffResponse struct {
	*flamebearer.FlamebearerProfile
	Metadata renderMetadataResponse `json:"metadata"`
}

func (p *provider) renderDiff(rw http.ResponseWriter, r *http.Request) {
	var params diffParams
	ctx := r.Context()

	if err := p.parseDiffQueryParams(r, &params); err != nil {
		httpserver.WriteErr(rw, "400", err.Error())
		return
	}

	leftOut, err := p.loadTree(ctx, &params.Left, params.Left.StartTime, params.Left.EndTime)
	if err != nil {
		httpserver.WriteErr(rw, "400", fmt.Sprintf("failed to load left tree: %w", err))
		return
	}

	rightOut, err := p.loadTree(ctx, &params.Right, params.Right.StartTime, params.Right.EndTime)
	if err != nil {
		httpserver.WriteErr(rw, "400", fmt.Sprintf("failed to load right tree: %w", err))
		return
	}

	leftProfile := flamebearer.ProfileConfig{
		Name:     "diff",
		MaxNodes: params.MaxNodes,
		Metadata: metadata.Metadata{
			SpyName:    leftOut.SpyName,
			SampleRate: leftOut.SampleRate,
			Units:      leftOut.Units,
		},
		Tree:      leftOut.Tree,
		Timeline:  leftOut.Timeline,
		Groups:    leftOut.Groups,
		Telemetry: leftOut.Telemetry,
	}

	rightProfile := flamebearer.ProfileConfig{
		Name:     "diff",
		MaxNodes: params.MaxNodes,
		Metadata: metadata.Metadata{
			SpyName:    rightOut.SpyName,
			SampleRate: rightOut.SampleRate,
			Units:      rightOut.Units,
		},
		Tree:      rightOut.Tree,
		Timeline:  rightOut.Timeline,
		Groups:    rightOut.Groups,
		Telemetry: rightOut.Telemetry,
	}

	combined, err := flamebearer.NewCombinedProfile(leftProfile, rightProfile)
	if err != nil {
		httpserver.WriteErr(rw, "400", fmt.Sprintf("failed to create combined profile: %w", err))
		return
	}

	switch params.Format {
	case "json":
		fallthrough
	default:
		md := renderMetadataResponse{FlamebearerMetadataV1: combined.Metadata}
		p.enhanceWithCustomFields(&md, params)

		res := RenderDiffResponse{
			FlamebearerProfile: &combined,
			Metadata:           md,
		}
		httpserver.WriteData(rw, res)
	}
}

func (p *provider) loadTree(ctx context.Context, gi *storage.GetInput, startTime, endTime time.Time) (_ *storage.GetOutput, _err error) {
	defer func() {
		rerr := recover()
		if rerr != nil {
			_err = fmt.Errorf("panic: %v", rerr)
			p.Log.Errorf("loadTree: recovered from panic: %v", rerr)
		}
	}()

	_gi := *gi
	_gi.StartTime = startTime
	_gi.EndTime = endTime
	out, err := p.st.Get(ctx, &_gi)
	if err != nil {
		return nil, err
	}
	if out == nil {
		return &storage.GetOutput{Tree: tree.New()}, nil
	}
	return out, nil
}

// add custom fields to renderMetadataResponse
// original motivation is to add custom {start,end}Time calculated dynamically
func (p *provider) enhanceWithCustomFields(md *renderMetadataResponse, params diffParams) {
	var diffAppName string

	if params.Left.Query.AppName == params.Right.Query.AppName {
		diffAppName = fmt.Sprintf("diff_%s_%s", params.Left.Query.AppName, params.Right.Query.AppName)
	} else {
		diffAppName = fmt.Sprintf("diff_%s", params.Left.Query.AppName)
	}

	startTime, endTime := p.findStartEndTime(params.Left, params.Right)

	md.AppName = diffAppName
	md.StartTime = startTime.Unix()
	md.EndTime = endTime.Unix()
	// TODO: add missing fields
}

// parseDiffQueryParams parses query params into a diffParams
func (p *provider) parseDiffQueryParams(r *http.Request, req *diffParams) (err error) {
	parseDiffQueryParams := func(r *http.Request, prefix string) (gi storage.GetInput, err error) {
		v := r.URL.Query()
		getWithPrefix := func(param string) string {
			return v.Get(prefix + strings.Title(param))
		}

		// Parse query
		qry, err := flameql.ParseQuery(getWithPrefix("query"))
		if err != nil {
			return gi, fmt.Errorf("%q: %+w", "Error parsing query", err)
		}
		gi.Query = qry

		gi.StartTime = attime.Parse(getWithPrefix("from"))
		gi.EndTime = attime.Parse(getWithPrefix("until"))

		return gi, nil
	}

	req.Left, err = parseDiffQueryParams(r, "left")
	if err != nil {
		return fmt.Errorf("%q: %+w", "Could not parse 'left' side", err)
	}

	req.Right, err = parseDiffQueryParams(r, "right")
	if err != nil {
		return fmt.Errorf("%q: %+w", "Could not parse 'right' side", err)
	}

	// Parse the common fields
	v := r.URL.Query()
	req.MaxNodes = p.Cfg.MaxNodesRender
	if mn, err := strconv.Atoi(v.Get("max-nodes")); err == nil && mn != 0 {
		req.MaxNodes = mn
	}

	req.Format = v.Get("format")
	return expectFormats(req.Format)
}

func (*provider) findStartEndTime(left storage.GetInput, right storage.GetInput) (time.Time, time.Time) {
	startTime := left.StartTime
	if right.StartTime.Before(left.StartTime) {
		startTime = right.StartTime
	}

	endTime := left.EndTime
	if right.EndTime.After(right.EndTime) {
		endTime = right.EndTime
	}

	return startTime, endTime
}
