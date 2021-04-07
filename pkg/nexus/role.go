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

package nexus

import (
	"bytes"
	"encoding/json"
	"net/url"
)

type RoleID string
type ExternalRoleID string

type Role struct {
	ID          string        `json:"id"`
	Source      string        `json:"source"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Privileges  []PrivilegeID `json:"privileges"`
	Roles       []RoleID      `json:"roles"`
}

type RoleListRequest struct {
	// +optional
	Source string
}

type RoleCreateRequest struct {
	// Role-ID must be unique
	ID          RoleID `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	// The list of privileges assigned to this role.
	Privileges []PrivilegeID `json:"privileges"`
	// The list of roles assigned to this role.
	Roles []string `json:"roles"`
}

type RoleGetRequest struct {
	ID     RoleID
	Source UserSource
}

type RoleUpdateRequest RoleCreateRequest

type RoleDeleteRequest struct {
	ID string
}

//////////////////////////////////////////
// http client
//////////////////////////////////////////

func (n *Nexus) EnsureRole(req RoleCreateRequest) error {
	_, err := n.GetRole(RoleGetRequest{ID: req.ID, Source: UserSourceDefault})
	if err != nil {
		if err != ErrNotFound {
			return err
		}
		// not found, create
		return n.CreateRole(req)
	}
	// update
	return n.UpdateRole(RoleUpdateRequest(req))
}

func (n *Nexus) CreateRole(req RoleCreateRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Post(n.Addr).Path("/service/rest/beta/security/roles").
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		JSONBody(req).
		Do().Body(&body)
	if err != nil {
		return err
	}
	if !httpResp.IsOK() {
		return ErrNotOK(httpResp.StatusCode(), body.String())
	}

	return nil
}

func (n *Nexus) UpdateRole(req RoleUpdateRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Put(n.Addr).Path("/service/rest/beta/security/roles/"+string(req.ID)).
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		JSONBody(req).
		Do().Body(&body)
	if err != nil {
		return err
	}
	if !httpResp.IsOK() {
		return ErrNotOK(httpResp.StatusCode(), body.String())
	}

	return nil
}

func (n *Nexus) DeleteRole(req RoleDeleteRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Delete(n.Addr).Path("/service/rest/beta/security/roles/"+req.ID).
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		Do().Body(&body)
	if err != nil {
		return err
	}
	if !httpResp.IsOK() {
		return ErrNotOK(httpResp.StatusCode(), body.String())
	}

	return nil
}

func (n *Nexus) GetRole(req RoleGetRequest) (*Role, error) {
	if req.Source == "" {
		req.Source = UserSourceDefault
	}

	var body bytes.Buffer
	httpResp, err := n.hc.Get(n.Addr).Path("/service/rest/beta/security/roles/"+string(req.ID)).
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		Param("source", string(req.Source)).
		Do().Body(&body)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, ErrNotOK(httpResp.StatusCode(), body.String())
	}

	var role Role
	if err := json.NewDecoder(&body).Decode(&role); err != nil {
		return nil, err
	}

	return &role, nil
}

func (n *Nexus) ListRoles(req RoleListRequest) ([]Role, error) {
	params := make(url.Values)
	if req.Source != "" {
		params["source"] = []string{req.Source}
	}

	var body bytes.Buffer
	httpResp, err := n.hc.Get(n.Addr).Path("/service/rest/beta/security/roles").
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		Params(params).
		Do().Body(&body)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, ErrNotOK(httpResp.StatusCode(), body.String())
	}

	var roles []Role
	if err := json.NewDecoder(&body).Decode(&roles); err != nil {
		return nil, err
	}

	return roles, nil
}
