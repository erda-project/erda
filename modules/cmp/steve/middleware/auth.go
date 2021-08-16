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

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	apiuser "k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/steve"
)

const varsKey = "stevePathVars"

type Authenticator struct {
	bdl *bundle.Bundle
}

// NewAuthenticator return a steve Authenticator with bundle.
// bdl need withCoreServices to check permission.
func NewAuthenticator(bdl *bundle.Bundle) *Authenticator {
	return &Authenticator{bdl: bdl}
}

// AuthMiddleware authenticate for steve server by bundle.
func (a *Authenticator) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		vars := parseVars(req)
		ctx := context.WithValue(req.Context(), varsKey, vars)
		req = req.WithContext(ctx)
		clusterName := vars["clusterName"]
		typ := vars["type"]
		userID := req.Header.Get("User-ID")
		orgID := req.Header.Get("Org-ID")
		scopeID, err := strconv.ParseUint(orgID, 10, 64)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write(apistructs.NewSteveError(apistructs.BadRequest, "invalid org id").JSON())
			return
		}

		clusters, err := a.bdl.ListClusters("k8s", scopeID)
		if err != nil {
			logrus.Errorf("failed to list cluster %s in steve authenticate, %v", clusterName, err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write(apistructs.NewSteveError(apistructs.ServerError, "check permission internal error").JSON())
			return
		}

		found := false
		for _, cluster := range clusters {
			if cluster.Name == clusterName {
				found = true
				break
			}
		}
		if !found {
			resp.WriteHeader(http.StatusNotFound)
			resp.Write(apistructs.NewSteveError(apistructs.NotFound,
				fmt.Sprintf("cluster %s not found in target org", clusterName)).JSON())
			return
		}

		r := &apistructs.ScopeRoleAccessRequest{
			Scope: apistructs.Scope{
				Type: apistructs.OrgScope,
				ID:   orgID,
			},
		}
		rsp, err := a.bdl.ScopeRoleAccess(userID, r)
		if err != nil {
			logrus.Errorf("failed to get scope role access in steve authentication, %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write(apistructs.NewSteveError(apistructs.ServerError, "check permission internal error").JSON())
			return
		}
		if !rsp.Access {
			resp.WriteHeader(http.StatusForbidden)
			resp.Write(apistructs.NewSteveError(apistructs.PermissionDenied, "access denied").JSON())
			return
		}

		name := fmt.Sprintf("erda-user-%s", userID)
		user := &apiuser.DefaultInfo{
			Name: name,
			UID:  name,
		}
		for _, role := range rsp.Roles {
			if role == bundle.RoleOrgManager {
				user.Groups = append(user.Groups, steve.OrgManagerGroup)
			}
			if role == bundle.RoleOrgSupport {
				user.Groups = append(user.Groups, steve.OrgSupportGroup)
			}
		}

		if len(user.Groups) == 0 {
			resp.WriteHeader(http.StatusForbidden)
			resp.Write(apistructs.NewSteveError(apistructs.PermissionDenied, "access denied").JSON())
			return
		}

		if req.Method == http.MethodGet && typ == "nodes" {
			user = &apiuser.DefaultInfo{
				Name: "admin",
				UID:  "admin",
				Groups: []string{
					"system:masters",
					"system:authenticated",
				},
			}
		}

		ctx = request.WithUser(ctx, user)
		req = req.WithContext(ctx)
		next.ServeHTTP(resp, req)
	})
}

func parseVars(req *http.Request) map[string]string {
	var match mux.RouteMatch
	m := mux.NewRouter().PathPrefix("/api/k8s/clusters/{clusterName}")
	s := m.Subrouter()
	s.Path("/v1")
	s.Path("/v1/")
	s.Path("/v1/{type}")
	s.Path("/v1/{type}/{name}")
	s.Path("/v1/{type}/{namespace}/{name}")
	s.Path("/v1/{type}/{namespace}/{name}/{link}")
	s.Path("/api/{version}/namespaces/{namespace}/{type}")
	s.Path("/api/{version}/namespaces/{namespace}/{type}/{name}")
	s.Path("/api/{version}/namespaces/{namespace}/{type}/{name}/{link}")

	vars := make(map[string]string)
	if s.Match(req, &match) {
		vars = match.Vars
	}

	var shellMatch mux.RouteMatch
	shellRouter := mux.NewRouter().Path("/api/k8s/clusters/{clusterName}/kubectl-shell")
	if shellRouter.Match(req, &shellMatch) {
		vars = shellMatch.Vars
		vars["kubectl-shell"] = "true"
	}
	return vars
}
