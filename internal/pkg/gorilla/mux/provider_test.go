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

package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda-infra/base/logs/logrusx"
)

func newTestProvider(t *testing.T) *provider {
	t.Helper()
	p := &provider{
		Config: &Config{},
		L:      logrusx.New(),
	}
	require.NoError(t, p.Init(nil))
	return p
}

func TestForceHandleReplacesExistingHandler(t *testing.T) {
	p := newTestProvider(t)

	const path = "/force"
	var firstHandlerHits int
	p.Handle(path, http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHandlerHits++
		w.WriteHeader(http.StatusAccepted)
	}))

	// Ensure the initial handler is registered and usable.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	p.Router.ServeHTTP(rec, req)
	require.Equal(t, 1, firstHandlerHits)
	require.Equal(t, http.StatusAccepted, rec.Code)

	var (
		secondHandlerHits int
	)
	p.ForceHandle(path, http.MethodGet, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHandlerHits++
		w.WriteHeader(http.StatusTeapot)
	}))

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, path, nil)
	p.Router.ServeHTTP(rec, req)

	require.Equal(t, 1, firstHandlerHits, "first handler should not receive requests after replacement")
	require.Equal(t, 1, secondHandlerHits)
	require.Equal(t, http.StatusTeapot, rec.Code)
}

func TestForceHandleRegistersWhenMissing(t *testing.T) {
	p := newTestProvider(t)

	const path = "/force-new"
	var hits int
	p.ForceHandle(path, "get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	p.Router.ServeHTTP(rec, req)

	require.Equal(t, 1, hits)
	require.Equal(t, http.StatusOK, rec.Code)
}
