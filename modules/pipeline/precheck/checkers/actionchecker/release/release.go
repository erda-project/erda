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

package release

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

const (
	paramCrossCluster = "cross_cluster"
	ActionType        = "release"
)

type release struct{}

func New() *release {
	return &release{}
}

func (r *release) ActionType() pipelineyml.ActionType {
	return ActionType
}

func (r *release) Check(ctx context.Context, data interface{}, itemsForCheck prechecktype.ItemsForCheck) (abort bool, messages []string) {
	// data type: pipelineyml.Action
	actualAction, ok := data.(pipelineyml.Action)
	if !ok {
		abort = false
		return
	}

	// check dice.yml
	diceymlContent, _ := itemsForCheck.Files["dice.yml"]
	if diceymlContent == "" {
		abort = false
		return
	}

	checkDiceYml := true
	if actualAction.Params != nil {
		checkDiceYmlStr, ok := actualAction.Params["check_diceyml"]
		if ok {
			b, err := strconv.ParseBool(checkDiceYmlStr.(string))
			if err != nil {
				abort = true
				messages = append(messages, fmt.Sprintf("invalid param 'check_diceyml', value: %s", checkDiceYmlStr))
			}
			checkDiceYml = b
		}
	}

	// parse dice.yml
	d, err := diceyml.New([]byte(diceymlContent), checkDiceYml)
	if err != nil {
		abort = true
		messages = append(messages, fmt.Sprintf("failed to parse dice.yml, err: %v", err))
		return
	}
	// check images
	if actualAction.Params != nil {
		// param: image
		imageJson, ok := actualAction.Params["image"]
		if ok {
			images := make(map[string]string)
			err := json.Unmarshal([]byte(imageJson.(string)), &images)
			if err != nil {
				abort = true
				messages = append(messages, fmt.Sprintf("invalid param image, err: %v", err))
			}
			if err := d.InsertImage(images, nil); err != nil {
				abort = true
				messages = append(messages, fmt.Sprintf("failed to insert image into dice.yml, err: %v", err))
				return
			}
		}
	}

	// check cross_cluster 参数
	spec := itemsForCheck.ActionSpecs[actualAction.GetActionTypeVersion()]
	crossCluster := false
	for _, param := range spec.Params {
		if param.Name == paramCrossCluster {
			v, ok := actualAction.Params[paramCrossCluster]
			if !ok {
				// use default value
				v = param.Default
			}
			switch v.(type) {
			case bool:
				crossCluster = v.(bool)
			case string:
				crossCluster, _ = strconv.ParseBool(v.(string))
			}
		}
	}
	prechecktype.PutContextResult(ctx, prechecktype.CtxResultKeyCrossCluster, crossCluster)

	return
}
