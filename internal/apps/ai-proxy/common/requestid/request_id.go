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

package requestid

import (
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func GetOrSetRequestID(r *http.Request) string {
	return getOrSetID(r, vars.XRequestId)
}

// GetCallID always generate a new uuid.
func GetCallID(r *http.Request) string {
	return getOrSetID(r, "")
}

func getOrSetID(r *http.Request, headerKey string) string {
	v := r.Header.Get(headerKey)
	if v != "" {
		return v
	}
	v = uuid.New()
	if headerKey != "" {
		r.Header.Set(headerKey, v)
	}
	return v
}
