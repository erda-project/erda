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
	"net/http"

	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/util/attime"

	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (p *provider) getLabels(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v := r.URL.Query()

	in := storage.GetLabelKeysByQueryInput{
		StartTime: attime.Parse(v.Get("from")),
		EndTime:   attime.Parse(v.Get("until")),
		Query:     v.Get("query"),
	}

	keys := make([]string, 0)
	if in.Query != "" {
		output, err := p.st.GetKeysByQuery(ctx, in)
		if err != nil {
			httpserver.WriteErr(rw, "400", err.Error())
			return
		}
		keys = append(keys, output.Keys...)
	} else {
		p.st.GetKeys(ctx, func(k string) bool {
			keys = append(keys, k)
			return true
		})
	}

	httpserver.WriteData(rw, keys)
}

func (p *provider) getLabelValues(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	v := r.URL.Query()

	in := storage.GetLabelValuesByQueryInput{
		StartTime: attime.Parse(v.Get("from")),
		EndTime:   attime.Parse(v.Get("until")),
		Label:     v.Get("label"),
		Query:     v.Get("query"),
	}

	if in.Label == "" {
		httpserver.WriteErr(rw, "400", "label parameter is required")
		return
	}

	values := make([]string, 0)
	if in.Query != "" {
		output, err := p.st.GetValuesByQuery(ctx, in)
		if err != nil {
			httpserver.WriteErr(rw, "400", err.Error())
			return
		}
		values = append(values, output.Values...)
	} else {
		p.st.GetValues(ctx, in.Label, func(v string) bool {
			values = append(values, v)
			return true
		})
	}
	httpserver.WriteData(rw, values)
}
