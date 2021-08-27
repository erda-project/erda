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
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	_ "github.com/erda-project/erda-proto-go" // import all protobuf APIs
	common "github.com/erda-project/erda-proto-go/common/pb"
)

// RangeOpenAPIsProxy .
func RangeOpenAPIsProxy(handler func(method, publishPath, backendPath, serviceName string, opt *common.OpenAPIOption) error) error {
	return RangeOpenAPIs("erda.", func(serviceName, method, backendPath string, opt *common.OpenAPIOption) error {
		if opt.Private {
			return nil
		}
		service := opt.Service
		if len(service) <= 0 {
			service = serviceName
		}

		backPrefix := opt.BackendPrefix
		if len(backPrefix) > 0 {
			backPrefix = formatPath(backPrefix)
			if !strings.HasPrefix(backendPath, backPrefix) {
				return fmt.Errorf("backend path %q must has prefix %q", backendPath, backPrefix)
			}
		}

		publishPath := opt.Path
		if len(publishPath) > 0 {
			publishPath = formatPath(publishPath)
		} else if len(backPrefix) > 0 {
			publishPath = formatPath(backendPath[len(backPrefix):])
		} else {
			publishPath = backendPath
		}

		prefix := opt.Prefix
		if len(prefix) > 0 {
			publishPath = strings.TrimRight(prefix, "/") + "/" + strings.TrimLeft(publishPath, "/")
		}

		return handler(method, publishPath, backendPath, service, opt)
	})
}

// RangeOpenAPIs .
func RangeOpenAPIs(pkgPrefix string, handler func(serviceName, method, path string, opt *common.OpenAPIOption) error) (err error) {
	files := protoregistry.GlobalFiles
	files.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		pkgName := string(fd.Package())
		if !strings.HasPrefix(pkgName, pkgPrefix) {
			return true
		}
		services := fd.Services()
		for i, n := 0, services.Len(); i < n; i++ {
			service := services.Get(i)
			serviceName := string(service.FullName())
			serviceOption, _ := proto.GetExtension(service.Options(), common.E_OpenapiService).(*common.OpenAPIOption)
			methods := service.Methods()
			for i, n := 0, methods.Len(); i < n; i++ {
				method := methods.Get(i)
				if method.IsStreamingClient() || method.IsStreamingServer() {
					continue
				}
				methodOption, _ := proto.GetExtension(method.Options(), common.E_Openapi).(*common.OpenAPIOption)
				opt := getOpenAPIOption(serviceOption, methodOption)
				if opt == nil {
					continue
				}
				rule, ok := proto.GetExtension(method.Options(), annotations.E_Http).(*annotations.HttpRule)
				if rule != nil && ok {
					for _, bind := range rule.AdditionalBindings {
						method, path := getPathMethod(service, method, bind)
						if err = handler(serviceName, method, path, opt); err != nil {
							return false
						}
					}
					method, path := getPathMethod(service, method, rule)
					if err = handler(serviceName, method, path, opt); err != nil {
						return false
					}
				}
			}
		}
		return true
	})
	return err
}

func getOpenAPIOption(opts ...*common.OpenAPIOption) *common.OpenAPIOption {
	var opt *common.OpenAPIOption
	for _, o := range opts {
		if o == nil {
			continue
		}
		if opt == nil {
			opt = &common.OpenAPIOption{}
		}
		if len(o.Path) > 0 {
			opt.Path = o.Path
		}
		if len(o.Prefix) > 0 {
			opt.Prefix = o.Prefix
		}
		if len(o.BackendPrefix) > 0 {
			opt.BackendPrefix = o.BackendPrefix
		}
		if len(o.Service) > 0 {
			opt.Service = o.Service
		}
		if o.Private {
			opt.Private = o.Private
		}
	}
	return opt
}

func getPathMethod(service protoreflect.ServiceDescriptor, m protoreflect.MethodDescriptor, rule *annotations.HttpRule) (method, path string) {
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	if len(path) <= 0 {
		path = fmt.Sprintf("/%s/%s", service.FullName(), m.Name())
	} else {
		path = formatPath(path)
	}
	if len(method) <= 0 {
		method = "POST"
	}
	return
}

func formatPath(path string) string {
	return "/" + strings.TrimLeft(path, "/")
}
