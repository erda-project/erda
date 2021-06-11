// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
	transhttp "github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/apistructs"
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
	ScopeSys          = "sys"
	ScopeOrg          = "org"
	ScopeProject      = "project"
	ScopeApp          = "app"
	ScopePublisher    = "publisher"
	ScopeMicroService = "micro_service"
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

				// TODO: get userID from http or grpc
				var userID string
				httpReq := transhttp.ContextRequest(ctx)
				if httpReq != nil {
					userID = httpReq.Header.Get("User-ID")
				}

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
func NoPermMethod(method string) *Permission {
	return &Permission{method: method}
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
	return func(ctx context.Context, req interface{}) (string, error) {
		if req == nil {
			return "", fmt.Errorf("not found id for permission")
		}
		val := reflect.ValueOf(req)
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() != reflect.Struct {
			return "", fmt.Errorf("invalid request type")
		}
		val = val.FieldByName(field)
		if !val.IsValid() {
			return "", fmt.Errorf("not found id for permission")
		}
		v := val.Interface()
		if v == nil {
			return "", fmt.Errorf("not found id for permission")
		}
		return fmt.Sprint(v), nil
	}
}

// FiexdValue .
func FiexdValue(v string) ValueGetter {
	return func(ctx context.Context, req interface{}) (string, error) {
		return v, nil
	}
}
