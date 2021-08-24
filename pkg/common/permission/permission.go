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

package permission

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

// Interface .
type Interface interface {
	Check(perms ...*Permission) transport.ServiceOption
}

// Scope .
type Scope string

// Scope values
const (
	ScopeSys                   = "sys"
	ScopeOrg                   = "org"
	ScopeProject               = "project"
	ScopeApp                   = "app"
	ScopePublisher             = "publisher"
	ScopeMicroService          = "micro_service"
	MonitorProjectAlert string = "monitor_project_alert"
)

// Action .
type Action string

// Action values
const (
	ActionCreate  = "CREATE"
	ActionDelete  = "DELETE"
	ActionUpdate  = "UPDATE"
	ActionGet     = "GET"
	ActionList    = "LIST"
	ActionOperate = "OPERATE"
)

// Permission .
type Permission struct {
	method   string
	scope    ValueGetter
	resource ValueGetter
	action   Action
	id       ValueGetter
	// for String
	originalMethod   interface{}
	originalScope    interface{}
	originalResource interface{}
}

func (p *Permission) String() string {
	return fmt.Sprintf("%s {scope=%v resource=%v action=%s}", getMethodFullName(p.originalMethod), p.originalScope, p.originalResource, p.action)
}

// ValueGetter .
type ValueGetter func(ctx context.Context, req interface{}) (string, error)

func (p *provider) Check(perms ...*Permission) transport.ServiceOption {
	methods := make(map[string]*Permission)
	for _, perm := range perms {
		if _, ok := methods[perm.method]; ok {
			panic(fmt.Errorf("method %q already exists", perm.method))
		}
		if len(perm.method) <= 0 {
			panic(fmt.Errorf("invalid method %q", perm.method))
		}
		methods[perm.method] = perm
		if perm.resource != nil {
			p.Log.Infof("permission: %s\n", perm)
		}
	}
	return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		if p.Cfg.Skip {
			return h
		}
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			info := transport.ContextServiceInfo(ctx)
			perm := methods[info.Method()]
			if perm == nil {
				return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), "", "permission undefined")
			}
			if perm.resource != nil {
				scope, err := perm.scope(ctx, req)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), err.Error())
				}
				resource, err := perm.resource(ctx, req)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), err.Error())
				}
				id, err := perm.id(ctx, req)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), err.Error())
				}
				idval, err := strconv.ParseUint(id, 10, 64)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), fmt.Sprintf("invalid %s id=%q", scope, id))
				}

				userID := apis.GetUserID(ctx)
				resp, err := p.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
					UserID:   userID,
					Scope:    apistructs.ScopeType(scope),
					ScopeID:  idval,
					Resource: resource,
					Action:   string(perm.action),
				})
				if err != nil {
					return nil, errors.NewServiceInvokingError("CheckPermission", err)
				}
				if !resp.Access {
					return nil, errors.NewPermissionError(fmt.Sprintf("user/%s/%s/%s/resource/%s", userID, scope, id, resource), string(perm.action), "")
				}
			}
			return h(ctx, req)
		}
	})
}

// Method .
func Method(method interface{}, scope, resource interface{}, action Action, id ValueGetter) *Permission {
	return &Permission{
		method:           getMethodName(method),
		scope:            toValueGetter(scope),
		resource:         toValueGetter(resource),
		action:           action,
		id:               id,
		originalMethod:   method,
		originalScope:    scope,
		originalResource: resource,
	}
}

// NoPermMethod 。
func NoPermMethod(method interface{}) *Permission {
	return &Permission{method: getMethodName(method)}
}

func getMethodFullName(method interface{}) string {
	if method == nil {
		return ""
	}
	name, ok := method.(string)
	if ok {
		return name
	}
	val := reflect.ValueOf(method)
	if val.Kind() != reflect.Func {
		panic(fmt.Errorf("method %V not function", method))
	}
	fn := runtime.FuncForPC(val.Pointer())
	return fn.Name()
}

func getMethodName(method interface{}) string {
	if method == nil {
		return ""
	}
	name, ok := method.(string)
	if ok {
		return name
	}
	name = getMethodFullName(method)
	parts := strings.Split(name, ".")
	if len(parts) < 2 {
		panic(fmt.Errorf("function %s is not method of type", name))
	}
	name = parts[len(parts)-1]
	idx := strings.IndexFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_'
	})
	if idx >= 0 {
		return name[:idx]
	}
	return name
}

func toValueGetter(v interface{}) ValueGetter {
	switch v := v.(type) {
	case string:
		return FiexdValue(v)
	case ValueGetter:
		return v
	case nil:
		return nil
	default:
		if reflect.TypeOf(v).Kind() == reflect.String {
			return FiexdValue(fmt.Sprint(v))
		}
	}
	panic(fmt.Errorf("invalid value getter %V", v))
}

// FieldValue .
func FieldValue(field string) ValueGetter {
	fields := strings.Split(field, ".")
	last := len(fields) - 1
	return func(ctx context.Context, req interface{}) (string, error) {
		if value := req; value != nil {
			for i, field := range fields {
				val := reflect.ValueOf(value)
				for val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				if val.Kind() != reflect.Struct {
					return "", fmt.Errorf("invalid request type")
				}
				val = val.FieldByName(field)
				if !val.IsValid() {
					break
				}
				value = val.Interface()
				if value == nil {
					break
				}
				if i == last {
					return fmt.Sprint(value), nil
				}
			}
		}
		return "", fmt.Errorf("not found id for permission")
	}
}

// FiexdValue .
func FiexdValue(v string) ValueGetter {
	return func(ctx context.Context, req interface{}) (string, error) {
		return v, nil
	}
}

// HeaderValue
func HeaderValue(key string) ValueGetter {
	return func(ctx context.Context, req interface{}) (string, error) {
		header := transport.ContextHeader(ctx)
		if header != nil {
			for _, v := range header.Get(key) {
				if len(v) > 0 {
					return v, nil
				}
			}
		}
		return "", fmt.Errorf("not found id for permission")
	}
}

// OrgIDValue
func OrgIDValue() ValueGetter {
	return HeaderValue("org-id")
}
