package runtime

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

// ValueGetter .
type ValueGetter func(ctx context.Context, req interface{}) (string, error)

// Action .
type Action string

type Option func(*Permission)

type Permission struct {
	method                 string
	scope                  ValueGetter
	resource               ValueGetter
	action                 Action
	id                     ValueGetter
	skipPermInternalClient bool
	// for String
	originalMethod   interface{}
	originalScope    interface{}
	originalResource interface{}
}

func (p *provider) Check(perms ...*Permission) transport.ServiceOption {
	methods := make(map[string]*Permission)

	for _, perm := range perms {

		// 1、 获取到resource
		resource, _ := perm.resource(context.Background(), nil)

		// 2、 再判断有没有
		if _, ok := methods[perm.method+"-"+resource]; ok {
			panic(fmt.Errorf("method %q already exists for prermission", perm.method))
		}
		if len(perm.method) <= 0 {
			panic(fmt.Errorf("invalid method %q for prermission", perm.method))
		}

		// 3、 存入到map中
		methods[perm.method+"-"+resource] = perm
	}

	return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			info := transport.ContextServiceInfo(ctx)

			if info.Method() == "CountPRByWorkspace" {
				return h(ctx, req)
			}
			// 1、 获取到workspace， 如果没有就从请求体中获取到runtimeId，然后从数据库中查询到runtime，然后获取到workspace
			workspace := p.GetWorkspace(req)

			resource := GetRuntimeResource(workspace)

			perm := methods[info.Method()+"-"+resource]

			if perm == nil {
				return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), "", "permission undefined")
			}

			if perm.skipPermInternalClient && apis.IsInternalClient(ctx) {
				return h(ctx, req)
			}
			if perm != nil {
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
					if userID == "" {
						userID, err = p.FieldValue("Operator")(ctx, req)
					}

					fmt.Printf("Method: %s, CheckPermission userID: %s\n", info.Method(), userID)
					resp, err := p.Bundle.CheckPermission(&apistructs.PermissionCheckRequest{
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
			}

			return h(ctx, req)
		}
	})
}

// CheckRuntimeID 给请求体中携带runtimeID的接口鉴权使用的
func (p *provider) CheckRuntimeID(perms ...*Permission) transport.ServiceOption {
	methods := make(map[string]*Permission)

	for _, perm := range perms {

		// 1、 获取到resource
		resource, _ := perm.resource(context.Background(), nil)

		// 2、 再判断有没有
		if _, ok := methods[perm.method+"-"+resource]; ok {
			panic(fmt.Errorf("method %q already exists for prermission", perm.method))
		}
		if len(perm.method) <= 0 {
			panic(fmt.Errorf("invalid method %q for prermission", perm.method))
		}

		// 3、 存入到map中
		methods[perm.method+"-"+resource] = perm
	}

	return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			info := transport.ContextServiceInfo(ctx)

			runtimeID := p.GetRuntimeID(req)

			var runtime dbclient.Runtime
			runtimeIDInt, _ := strconv.ParseUint(runtimeID, 10, 64)
			if err := p.DB.Where("id = ?", runtimeIDInt).First(&runtime).Error; err != nil {
				return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), "", "permission undefined")
			}

			workspace := runtime.Workspace

			resource := GetRuntimeResource(workspace)

			perm := methods[info.Method()+"-"+resource]
			if perm.skipPermInternalClient && apis.IsInternalClient(ctx) {
				return h(ctx, req)
			}
			if perm != nil {
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
					userID := ""
					if info.Method() == "CreateRuntime" {
						userID, err = p.FieldValue("Operator")(ctx, req)
					} else {
						userID = apis.GetUserID(ctx)
					}
					fmt.Printf("Method: %s, CheckPermission userID: %s\n", info.Method(), userID)
					resp, err := p.Bundle.CheckPermission(&apistructs.PermissionCheckRequest{
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
			}

			return h(ctx, req)
		}
	})
}

// CheckRuntimeIDs 给请求体中携带runtimeIDs的接口鉴权使用的
func (p *provider) CheckRuntimeIDs(perms ...*Permission) transport.ServiceOption {
	methods := make(map[string]*Permission)

	for _, perm := range perms {

		// 1、 获取到resource
		resource, _ := perm.resource(context.Background(), nil)

		// 2、 再判断有没有
		if _, ok := methods[perm.method+"-"+resource]; ok {
			panic(fmt.Errorf("method %q already exists for prermission", perm.method))
		}
		if len(perm.method) <= 0 {
			panic(fmt.Errorf("invalid method %q for prermission", perm.method))
		}

		// 3、 存入到map中
		methods[perm.method+"-"+resource] = perm
	}

	return transport.WithInterceptors(func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			info := transport.ContextServiceInfo(ctx)

			workspaceStr, _ := p.FieldValue("Workspace")(ctx, req)
			if workspaceStr == "" {
				workspaceStr, _ = p.FieldValue("WorkSpace")(ctx, req)
			}

			// 去掉首尾的 []
			workspaceStr = strings.Trim(workspaceStr, "[]")
			if workspaceStr == "" {
				return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), "", "permission undefined")
			}

			workspace := strings.Split(workspaceStr, " ")[0]

			resource := GetRuntimeResource(workspace)

			perm := methods[info.Method()+"-"+resource]

			if perm.skipPermInternalClient && apis.IsInternalClient(ctx) {
				return h(ctx, req)
			}

			applicationStr, _ := perm.id(ctx, req)

			// 去掉首尾的 []
			applicationStr = strings.Trim(applicationStr, "[]")
			applicationIDs := strings.Split(applicationStr, " ")

			if perm != nil {
				scope, err := perm.scope(ctx, req)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), err.Error())
				}
				resource, err = perm.resource(ctx, req)
				if err != nil {
					return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), err.Error())
				}
				for _, applicationID := range applicationIDs {
					idval, err := strconv.ParseUint(applicationID, 10, 64)
					if err != nil {
						return nil, errors.NewPermissionError(info.Service()+"/"+info.Method(), string(perm.action), fmt.Sprintf("invalid %s id=%q", scope, applicationID))
					}
					userID := apis.GetUserID(ctx)
					resp, err := p.Bundle.CheckPermission(&apistructs.PermissionCheckRequest{
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
						return nil, errors.NewPermissionError(fmt.Sprintf("user/%s/%s/%s/resource/%s", userID, scope, applicationID, resource), string(perm.action), "")
					}

				}
			}

			return h(ctx, req)
		}
	})
}

func (p *provider) Method(method interface{}, scope, resource interface{}, action Action, id ValueGetter, options ...Option) *Permission {
	per := &Permission{
		method:           getMethodName(method),
		scope:            toValueGetter(scope),
		resource:         toValueGetter(resource),
		action:           action,
		id:               id,
		originalMethod:   method,
		originalScope:    scope,
		originalResource: resource,
	}
	for _, op := range options {
		op(per)
	}
	return per
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
	case func(ctx context.Context, req interface{}) (string, error):
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

func FiexdValue(v string) ValueGetter {
	return func(ctx context.Context, req interface{}) (string, error) {
		return v, nil
	}
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

// FieldValue 通过反射从请求体中取出值
func (p *provider) FieldValue(field string) func(ctx context.Context, req interface{}) (string, error) {
	fields := strings.Split(field, ".")
	last := len(fields) - 1

	result := func(ctx context.Context, req interface{}) (string, error) {
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
					valueStr := fmt.Sprint(value)
					return valueStr, nil
				}
			}
		}
		return "", fmt.Errorf("not found id for permission")
	}

	return result
}

// GetAppIDByRuntimeID 反射拿到runtimeID， 然后查数据库获取applicationID
func (p *provider) GetAppIDByRuntimeID(field string) func(ctx context.Context, req interface{}) (string, error) {
	fields := strings.Split(field, ".")

	result := func(ctx context.Context, req interface{}) (string, error) {
		if value := req; value != nil {
			for _, field := range fields {
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

				// 如果是ID，需要从数据库中获取到appID
				var runtime dbclient.Runtime
				if err := p.DB.Where("id = ?", value).First(&runtime).Error; err != nil {
					if gorm.IsRecordNotFoundError(err) {
						return "", nil
					}
				}
				appId := strconv.FormatUint(runtime.ApplicationID, 10)

				return appId, nil

			}
		}
		return "", fmt.Errorf("not found id for permission")
	}

	return result
}

// GetWorkspace 从请求体中获取workspace字段。
func (p *provider) GetWorkspace(req interface{}) string {
	ctx := context.Background()
	key_list := []string{"Workspace", "WorkSpace", "Extra.Workspace"}
	workspace := ""
	for _, key := range key_list {
		workspace, _ = p.FieldValue(key)(ctx, req)
		if workspace != "" {
			break
		}
	}
	return workspace
}

// GetRuntimeID 从请求体中获取 runtimeID
func (p *provider) GetRuntimeID(req interface{}) string {
	ctx := context.Background()
	runtimeId, err := p.FieldValue("RuntimeID")(ctx, req)
	if runtimeId == "" {
		runtimeId, err = p.FieldValue("NameOrID")(ctx, req)
	}
	if err != nil {
		return ""
	}
	return runtimeId
}

func GetRuntimeResource(workspace string) string {
	switch workspace {
	case "DEV":
		return "runtime-dev"
	case "[DEV]":
		return "runtime-dev"
	case "PROD":
		return "runtime-prod"
	case "[PROD]":
		return "runtime-prod"
	case "TEST":
		return "runtime-test"
	case "[TEST]":
		return "runtime-test"
	case "STAGING":
		return "runtime-staging"
	case "[STAGING]":
		return "runtime-staging"
	default:
		return ""
	}
}
