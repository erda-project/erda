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

package proto

import (
	"net/http"
	"reflect"

	common "github.com/erda-project/erda-proto-go/common/pb"
	auth "github.com/erda-project/erda/modules/core/openapi-ng/auth"
)

func getAuthOption(method, publishPath, backendPath, serviceName string, opt *common.OpenAPIOption) func(r *http.Request) auth.Options {
	return func(r *http.Request) auth.Options {
		return &authOption{opt: opt}
	}
}

type authOption struct {
	opt   *common.OpenAPIOption
	other map[string]interface{}
}

func (o *authOption) Get(key string) interface{} {
	if o.opt.Auth != nil {
		val := reflect.ValueOf(o.opt.Auth).FieldByName(key)
		if val.IsValid() && val.Kind() == reflect.Bool {
			return val.Bool()
		}
	}
	return o.other[key]
}

func (o *authOption) Set(key string, val interface{}) {
	if o.other == nil {
		o.other = make(map[string]interface{})
	}
	o.other[key] = val
}
