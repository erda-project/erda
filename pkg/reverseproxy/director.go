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

package reverseproxy

import (
	"context"
	"net/http"
	"sync"
)

func DoNothingDirector(_ *http.Request) {}

func AppendDirectors(ctx context.Context, director ...func(r *http.Request)) {
	var (
		m         = ctx.Value(CtxKeyMap{}).(*sync.Map)
		directors []func(*http.Request)
	)
	if value, ok := m.Load(MapKeyDirectors{}); ok && value != nil {
		if dirs, ok := value.([]func(r *http.Request)); ok {
			directors = dirs
		}
	}
	m.Store(MapKeyDirectors{}, append(directors, director...))
}
