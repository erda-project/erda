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
	"fmt"

	"github.com/erda-project/erda/pkg/strutil"
)

type PrivilegeID string

type Privilege struct {
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ReadOnly    bool     `json:"readOnly"`
	Pattern     string   `json:"pattern,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	Domain      string   `json:"domain,omitempty"`
}

type PrivilegeAction string

const (
	PrivilegeActionREAD         PrivilegeAction = "READ"
	PrivilegeActionBROWSE       PrivilegeAction = "BROWSE"
	PrivilegeActionEDIT         PrivilegeAction = "EDIT"
	PrivilegeActionADD          PrivilegeAction = "ADD"
	PrivilegeActionDELETE       PrivilegeAction = "DELETE"
	PrivilegeActionRUN          PrivilegeAction = "RUN"
	PrivilegeActionASSOCIATE    PrivilegeAction = "ASSOCIATE"
	PrivilegeActionDISASSOCIATE PrivilegeAction = "DISASSOCIATE"
	PrivilegeActionALL          PrivilegeAction = "ALL"
)

var (
	RepoDeploymentPrivileges = []PrivilegeAction{PrivilegeActionADD, PrivilegeActionBROWSE, PrivilegeActionEDIT, PrivilegeActionREAD}
	RepoReadOnlyPrivileges   = []PrivilegeAction{PrivilegeActionBROWSE, PrivilegeActionREAD}
)

type PrivilegeType string

const (
	PrivilegeTypeRepositoryView  PrivilegeType = "nx-repository-view"
	PrivilegeTypeRepositoryAdmin PrivilegeType = "nx-repository-admin"
)

type PrivilegeListRequest struct {
}

type PrivilegeGetRequest struct {
	PrivilegeID string
}

type PrivilegeDeleteRequest struct {
	PrivilegeID string
}

type RepositoryContentSelectorPrivilegeCreateRequest struct {
	// The name of the privilege. This value cannot be changed.
	// pattern: ^[a-zA-Z0-9\-]{1}[a-zA-Z0-9_\-\.]*$
	Name        string `json:"name"`
	Description string `json:"description"`
	// A collection of actions to associate with the privilege, using BREAD syntax (browse,read,edit,add,delete,all) as well as 'run' for script privileges.
	Actions []PrivilegeAction `json:"actions"`
	// The repository format (i.e 'nuget', 'npm') this privilege will grant access to (or * for all).
	Format RepositoryFormat `json:"format"`
	// The name of the repository this privilege will grant access to (or * for all).
	Repository string `json:"repository"`
	// The name of a content selector that will be used to grant access to content via this privilege.
	ContentSelector string `json:"contentSelector"`
}

type RepositoryContentSelectorPrivilegeUpdateRequest RepositoryContentSelectorPrivilegeCreateRequest

func GetNxRepositoryViewPrivileges(repoName string, repoFormat RepositoryFormat, actions ...PrivilegeAction) []PrivilegeID {
	var privileges []PrivilegeID
	for _, action := range actions {
		privileges = append(privileges,
			GetNxRepositoryPrivilege(PrivilegeTypeRepositoryView, repoFormat, repoName, action)...)
	}
	return privileges
}

func GetNxRepositoryPrivilege(privilegeType PrivilegeType, repoFormat RepositoryFormat, repoName string, actions ...PrivilegeAction) []PrivilegeID {
	var privilegeIDs []PrivilegeID
	for _, action := range actions {
		privilegeIDs = append(privilegeIDs,
			PrivilegeID(fmt.Sprintf("%s-%s-%s-%s", privilegeType, repoFormat.Standard(), repoName, strutil.ToLower(string(action)))))
	}
	return privilegeIDs
}

//////////////////////////////////////////
// http client
//////////////////////////////////////////

func (n *Nexus) ListPrivileges(req PrivilegeListRequest) ([]Privilege, error) {
	var body bytes.Buffer
	httpResp, err := n.hc.Get(n.Addr).Path("/service/rest/beta/security/privileges").
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		Do().Body(&body)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, ErrNotOK(httpResp.StatusCode(), body.String())
	}

	var privileges []Privilege
	if err := json.NewDecoder(&body).Decode(&privileges); err != nil {
		return nil, err
	}

	return privileges, nil
}

func (n *Nexus) GetPrivilege(req PrivilegeGetRequest) (*Privilege, error) {
	var body bytes.Buffer
	httpResp, err := n.hc.Get(n.Addr).Path("/service/rest/beta/security/privileges/"+req.PrivilegeID).
		Header(HeaderAuthorization, n.basicAuthBase64Value()).
		Do().Body(&body)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, ErrNotOK(httpResp.StatusCode(), body.String())
	}

	var privilege Privilege
	if err := json.NewDecoder(&body).Decode(&privilege); err != nil {
		return nil, err
	}

	return &privilege, nil
}

func (n *Nexus) DeletePrivilege(req PrivilegeDeleteRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Delete(n.Addr).Path("/service/rest/beta/security/privileges/"+req.PrivilegeID).
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

func (n *Nexus) CreateRepositoryContentSelectorPrivilege(req RepositoryContentSelectorPrivilegeCreateRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Post(n.Addr).Path("/service/rest/beta/security/privileges/repository-content-selector").
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

func (n *Nexus) UpdateRepositoryContentSelectorPrivilege(req RepositoryContentSelectorPrivilegeUpdateRequest) error {
	var body bytes.Buffer
	httpResp, err := n.hc.Put(n.Addr).Path("/service/rest/beta/security/privileges/repository-content-selector/"+req.Name).
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
