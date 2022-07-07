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

package httpserver

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Test_normalizePath(t *testing.T) {
	assert.Equal(t, "/api/projects/{}", normalizePath("/api/projects/{projectID}"))
	assert.Equal(t, "/api/projects/{}", normalizePath("/api/projects/{projectId}"))
}

func Test_genMapKeyForCompare(t *testing.T) {
	assert.Equal(t, "get:/api/projects/{}", genMapKeyForCompare(http.MethodGet, "/api/projects/{projectID}"))
	assert.Equal(t, "head:/api/projects/{}", genMapKeyForCompare(http.MethodHead, "/api/projects/{projectId}"))
}

func TestServer_CheckDuplicateRoutes(t *testing.T) {
	s := &Server{router: mux.NewRouter()}
	s.router.Path("/api/projects/{projectID}").Methods(http.MethodGet)
	s.router.Path("/api/projects/{projectId}").Methods(http.MethodGet, http.MethodPut)
	duplicates, ok := s.CheckDuplicateRoutes()
	assert.False(t, ok)
	assert.Equal(t, 1, len(duplicates))
	assert.Equal(t, "get:/api/projects/{}", duplicates[0])
}
