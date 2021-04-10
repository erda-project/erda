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

package configsetformmodal

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/pkg/strutil"
)

type ComponentFormModal struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type ConfigSetCreateForm struct {
	ClusterID     int64  `json:"cluster"`
	ConfigSetName string `json:"name"`
}

func (c *ComponentFormModal) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentFormModal) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

func (c *ComponentFormModal) OperateSubmit(orgID int64, identity apistructs.Identity) error {
	var (
		formEntity = ConfigSetCreateForm{}
	)

	jsonData, err := json.Marshal(c.component.State["formData"])
	if err != nil {
		return fmt.Errorf("marshal form data error: %v", err)
	}

	err = json.Unmarshal(jsonData, &formEntity)
	if err != nil {
		return fmt.Errorf("unmarshal form data error: %v", err)
	}

	err = validateSubmitData(formEntity)
	if err != nil {
		return fmt.Errorf("submit data error: %v", err)
	}

	req := &apistructs.EdgeConfigSetCreateRequest{
		OrgID:     orgID,
		Name:      formEntity.ConfigSetName,
		ClusterID: formEntity.ClusterID,
	}

	err = c.ctxBundle.Bdl.CreateEdgeConfigset(req, identity)
	if err != nil {
		return fmt.Errorf("create edge config set error: %v", err)
	}

	return nil
}

func validateSubmitData(formEntity ConfigSetCreateForm) error {
	if err := strutil.Validate(formEntity.ConfigSetName, strutil.MaxRuneCountValidator(apistructs.EdgeDefaultNameMaxLength)); err != nil {
		return err
	}

	isRight, err := regexp.MatchString(apistructs.EdgeDefaultMatchPattern, formEntity.ConfigSetName)
	if err != nil {
		return err
	}
	if !isRight {
		return fmt.Errorf(apistructs.EdgeDefaultRegexpError)
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFormModal{}
}
