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
)

type Mux interface {
	Handle(path, method string, h http.Handler, middles ...Middle)
	HandlePrefix(prefix, method string, h http.Handler, middles ...Middle)
	HandleMatch(match func(r *http.Request) bool, h http.Handler, middles ...Middle)
	HandleNotFound(h http.Handler, middles ...Middle)
	HandleMethodNotAllowed(h http.Handler, middles ...Middle)
}
