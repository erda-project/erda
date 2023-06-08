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

package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/erda-project/erda-infra/pkg/transport/http/runtime"
	common "github.com/erda-project/erda-proto-go/common/pb"
)

type Register interface {
	Register(route *APIProxy) error
}

// APIProxy .
type APIProxy struct {
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	ServiceURL  string          `json:"service_url"`
	BackendPath string          `json:"backend_path"`
	Auth        *common.APIAuth `json:"auth"`
}

// Validate .
func (a *APIProxy) Validate() error {
	a.Method = strings.TrimSpace(a.Method)
	a.Method = strings.ToUpper(a.Method)
	if a.Method == "" {
		a.Method = http.MethodGet
	}
	a.Path = strings.TrimSpace(a.Path)
	a.Path = "/" + strings.TrimLeft(a.Path, "/")
	a.BackendPath = strings.TrimSpace(a.BackendPath)
	a.BackendPath = "/" + strings.TrimLeft(a.BackendPath, "/")

	if len(a.Path) <= 0 {
		return fmt.Errorf("path is required")
	}
	if len(a.ServiceURL) <= 0 {
		return fmt.Errorf("service url is required")
	}
	u, err := url.Parse(a.ServiceURL)
	if err != nil {
		return fmt.Errorf("invalid service url: %w", err)
	}
	if len(u.Host) <= 0 {
		return fmt.Errorf("invalid service host is required")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("service url scheme must be http or https")
	}

	_, err = runtime.Compile(a.Path)
	if err != nil {
		return fmt.Errorf("invalid path %q: %s", a.Path, err)
	}
	_, err = runtime.Compile(a.BackendPath)
	if err != nil {
		return fmt.Errorf("invalid backend-path %q: %s", a.BackendPath, err)
	}
	return nil
}

func (a *APIProxy) Get(key string) any {
	// for "path", "method"
	field := reflect.ValueOf(a).Elem().FieldByName(key)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}

	// for "CheckLogin", "CheckToken" ...
	if a.Auth == nil {
		a.Auth = new(common.APIAuth)
	}
	field = reflect.ValueOf(a.Auth).Elem().FieldByName(key)
	if field.IsValid() || field.CanInterface() {
		return field.Interface()
	}

	return nil
}
