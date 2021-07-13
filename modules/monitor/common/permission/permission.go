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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	table "github.com/erda-project/erda/modules/monitor/common/db"
	bundlecmdb "github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	api "github.com/erda-project/erda/pkg/common/httpapi"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

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

// ValueGetter .
type ValueGetter func(ctx httpserver.Context) (string, error)

var (
	hc   = httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	bdl  *bundle.Bundle
	once sync.Once
	cmdb *bundlecmdb.Cmdb // 为了调用 /api/orgs/clusters/relations 接口
)

func initBundle() {
	bdl = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCoreServices(),
	)
	cmdb = bundlecmdb.New(bundlecmdb.WithHTTPClient(hc))
}

// Interceptor .
func Intercepter(scope interface{}, id ValueGetter, resource interface{}, action Action) httpserver.Interceptor {
	once.Do(initBundle)
	fmt.Printf("permission: scope = %s, resource = %s, action = %s\n", scope, resource, action)
	switch res := resource.(type) {
	case string:
		return check(scope, id, FiexdValue(res), action)
	case ValueGetter:
		return check(scope, id, res, action)
	case func(ctx httpserver.Context) (string, error):
		return check(scope, id, res, action)
	}
	panic(fmt.Errorf("invalid resource type: %v", resource))
}

func check(scope interface{}, idget ValueGetter, resget ValueGetter, action Action) httpserver.Interceptor {
	return func(handler func(ctx httpserver.Context) error) func(ctx httpserver.Context) error {
		return func(ctx httpserver.Context) error {
			req := ctx.Request()
			resource, err := resget(ctx)
			if err != nil {
				Failure(ctx, err.Error())
				return nil
			}
			id, err := idget(ctx)
			if err != nil {
				Failure(ctx, err.Error())
				return err
			}
			var scopeType apistructs.ScopeType
			switch res := scope.(type) {
			case string:
				scopeType = apistructs.ScopeType(res)
			case ValueGetter:
				f := res
				s, err := f(ctx)
				if err != nil {
					return nil
				}
				scopeType = apistructs.ScopeType(s)
			}
			idval, err := strconv.ParseUint(id, 10, 64)
			if err != nil {
				Failure(ctx, fmt.Sprintf("fail to convert scope id: %s", err))
				return nil
			}
			resp, err := bdl.CheckPermission(&apistructs.PermissionCheckRequest{
				UserID:   req.Header.Get("User-ID"),
				Scope:    scopeType,
				ScopeID:  idval,
				Resource: resource,
				Action:   string(action),
			})
			if err != nil {
				Failure(ctx, err.Error())
				return nil
			}
			if !resp.Access {
				Failure(ctx, fmt.Sprintf("user:%s, scope:%s/%s, res:%s, action:%s", req.Header.Get("User-ID"), scope, id, resource, action))
				return nil
			}
			return handler(ctx)
		}
	}
}

// Failure .
func Failure(context httpserver.Context, ctx interface{}) {
	resp := api.Errors.AccessDenied(ctx).Response(context)
	w := context.ResponseWriter()
	if resp.Status(context) > 0 {
		w.WriteHeader(resp.Status(context))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
	reader := resp.ReadCloser(context)
	io.Copy(context.ResponseWriter(), reader)
	reader.Close()
}

// QueryValue .
func QueryValue(keys ...string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		params := ctx.Request().URL.Query()
		for _, key := range keys {
			vals := params[key]
			if len(vals) == 1 {
				val := vals[0]
				if len(val) == 0 {
					continue
				}
				return val, nil
			}
			if len(vals) > 1 {
				return "", fmt.Errorf("too many key %s present", key)
			}
		}
		return "", fmt.Errorf("keys %v not found", keys)
	}
}

// PathValue .
func PathValue(keys ...string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		for _, key := range keys {
			val := ctx.Param(key)
			if len(val) == 0 {
				continue
			}
			return val, nil
		}
		return "", fmt.Errorf("keys %v not found", keys)
	}
}

// FiexdValue .
func FiexdValue(v string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return v, nil
	}
}

func projectIdFromScopeId(ctx httpserver.Context, db *table.DB) (string, error) {
	value := QueryValue("scopeId")
	tk, err := value(ctx)
	if err != nil {
		return "", err
	}
	projectId, err := db.Monitor.SelectProjectIdByTk(tk)
	if err != nil {
		return "", err
	}
	return projectId, nil
}

func projectIdFromTenantGroup(ctx httpserver.Context, db *table.DB) (string, error) {
	value := QueryValue("tenantGroup")
	tenantGroup, err := value(ctx)
	if err != nil {
		return "", err
	}
	tk, err := db.InstanceTenant.QueryTkByTenantGroup(tenantGroup)
	if err != nil {
		return "", err
	}
	projectId, err := db.Monitor.SelectProjectIdByTk(tk)
	if err != nil {
		return "", err
	}
	return projectId, nil
}

func ScopeIdFromParams(db *table.DB) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return projectIdFromScopeId(ctx, db)
	}
}

func TenantGroupFromParams(db *table.DB) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return projectIdFromTenantGroup(ctx, db)
	}
}

func TkFromParams(db *table.DB) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return tkFromParams(ctx, db)
	}
}

func ProjectIdFromParams() func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return projectIdFromParams(ctx)
	}
}

func projectIdFromParams(ctx httpserver.Context) (string, error) {
	projectId := ProjectId(ctx.Request())
	return projectId, nil
}

func tkFromParams(ctx httpserver.Context, db *table.DB) (string, error) {
	tk := Tk(ctx.Request())
	projectId, err := db.Monitor.SelectProjectIdByTk(tk)
	return projectId, err
}

// OrgIDFromHeader .
func OrgIDFromHeader() func(ctx httpserver.Context) (string, error) {
	return orgIDFromHeader
}

func orgIDFromHeader(ctx httpserver.Context) (string, error) {
	return api.OrgID(ctx.Request()), nil
}

// OrgIDFromQuery .
func OrgIDFromQuery(key string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		return ctx.Request().FormValue(key), nil
	}
}

// OrgIDByCluster .
func OrgIDByCluster(key string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		req := ctx.Request()
		idStr := api.OrgID(req)
		orgID, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("Org-ID is not number")
		}
		cluster := req.URL.Query().Get(key)
		if len(cluster) <= 0 {
			return "", fmt.Errorf("cluster must not be empty")
		}
		err = checkOrgIDsByCluster(orgID, cluster)
		if err != nil {
			return "", err
		}
		return idStr, nil
	}
}

type MonitorPermission struct {
	Name string
}

/*
	查询全部关联关系，再在内存中过滤
	TODO 待优化，目前只有这个接口
*/
func checkOrgIDsByCluster(orgID uint64, clusterName string) error {
	resp, err := cmdb.QueryAllOrgClusterRelation()
	if err != nil {
		return err
	}
	for _, item := range resp {
		if item.ClusterName == clusterName {
			if orgID == item.OrgID {
				return nil
			}
		}
	}
	return fmt.Errorf("not found cluster '%s'", clusterName)
}

// OrgIDByOrgName .
func OrgIDByOrgName(key string) func(ctx httpserver.Context) (string, error) {
	return func(ctx httpserver.Context) (string, error) {
		req := ctx.Request()
		name := req.URL.Query().Get(key)
		if len(name) <= 0 {
			return "", fmt.Errorf("not found org name from key %s", key)
		}
		orgInfo, err := bdl.GetOrg(name)
		if err != nil {
			return "", fmt.Errorf("fail to found org info: %s", err)
		}
		if orgInfo == nil {
			return "", fmt.Errorf("fail to found org info")
		}
		ctx.SetAttribute("Org-ID", int(orgInfo.ID))
		return strconv.Itoa(int(orgInfo.ID)), nil
	}
}

func ProjectId(r *http.Request) string {
	projectId := ""
	queries := strings.Split(r.URL.RawQuery, "&")

	for _, value := range queries {

		param := strings.Split(value, "=")
		k := param[0]
		v := param[1]

		if "projectId" == k || "project_id" == k {
			projectId = v
			break
		}
	}
	return projectId
}

func Tk(r *http.Request) string {
	tk := ""
	queries := strings.Split(r.URL.RawQuery, "&")

	for _, value := range queries {

		param := strings.Split(value, "=")
		k := param[0]
		v := param[1]

		if "terminusKey" == k || "terminus_key" == k {
			tk = v
			break
		}
	}
	return tk
}
