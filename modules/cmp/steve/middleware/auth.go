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
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

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
	return http.HandlerFunc(func(resp http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		clusterName := vars["clusterName"]

		userID := request.Header.Get("User-ID")
		if userID == "" {
			resp.WriteHeader(http.StatusUnauthorized)
			resp.Write([]byte("invalid user id\n"))
			return
		}

		orgID := request.Header.Get("Org-ID")
		scopeID, err := strconv.ParseUint(orgID, 10, 64)
		if err != nil {
			resp.WriteHeader(http.StatusUnauthorized)
			resp.Write([]byte("invalid org id\n"))
			return
		}

		clusters, err := a.bdl.ListClusters("k8s", scopeID)
		if err != nil {
			logrus.Errorf("failed to list cluster %s in steve authenticate, %v", clusterName, err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte("check permission internal error\n"))
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
			resp.Write([]byte(fmt.Sprintf("cluster %s not found in org %d\n", clusterName, scopeID)))
			return
		}

		action := ""
		switch request.Method {
		case http.MethodGet:
			if _, ok := vars["name"]; ok {
				action = "GET"
			} else {
				action = "LIST"
			}
		case http.MethodPost:
			action = "CREATE"
		case http.MethodPatch:
			fallthrough
		case http.MethodPut:
			action = "UPDATE"
		case http.MethodDelete:
			action = "DELETE"
		default:
			resp.WriteHeader(http.StatusMethodNotAllowed)
			resp.Write([]byte(fmt.Sprintf("method %s is not allowed\n", request.Method)))
			return
		}

		r := &apistructs.PermissionCheckRequest{
			UserID:   userID,
			Scope:    "org",
			ScopeID:  scopeID,
			Resource: "steve-api",
			Action:   action,
		}

		rsp, err := a.bdl.CheckPermission(r)
		if err != nil {
			logrus.Errorf("failed to check permission in steve authenticate, %v", err)
			resp.WriteHeader(http.StatusInternalServerError)
			resp.Write([]byte("check permission internal error\n"))
			return
		}

		if !rsp.Access {
			resp.WriteHeader(http.StatusMethodNotAllowed)
			resp.Write([]byte("access denied\n"))
			return
		}

		next.ServeHTTP(resp, request)
	})
}
