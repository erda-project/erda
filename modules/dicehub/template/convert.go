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

package template

import (
	"encoding/json"
	"errors"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/dicehub/template/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type TemplateVersion int

const (
	TemplateVersionV1 = TemplateVersion(1)
	TemplateVersionV2 = TemplateVersion(2)
)

func checkTemplateSpec(p *pb.PipelineTemplateSpec) error {
	if p.Name == "" {
		return errors.New("spec name can not empty")
	}

	if p.Template == "" {
		return errors.New("spec template can not empty")
	}

	if p.Params != nil {
		for _, v := range p.Params {
			if err := CheckPipelineParam(v); err != nil {
				return err
			}
		}
	}

	if p.Outputs != nil {
		for _, v := range p.Outputs {
			if err := CheckPipelineOutput(v); err != nil {
				return err
			}
		}
	}

	return nil
}

func CheckPipelineOutput(output *pb.PipelineOutput) error {
	if output.Name == "" {
		return errors.New("outputs name can not empty")
	}

	if output.Ref == "" {
		return errors.New("outputs ref can not empty")
	}

	return nil
}

func CheckPipelineParam(params *pb.PipelineParam) error {

	if params.Name == "" {
		return errors.New("params name can not empty")
	}

	return nil
}

func DoRenderTemplateWithFormatV2(params map[string]*structpb.Value, templateAction *pb.PipelineTemplateSpec, alias string, templateVersion TemplateVersion) (string, []*pb.SnippetFormatOutputs, error) {
	var (
		mp           map[string]interface{}
		templateSpec *apistructs.PipelineTemplateSpec
		respList     []*pb.SnippetFormatOutputs
	)
	mpData, err := json.Marshal(params)
	if err != nil {
		return "", nil, err
	}
	if err := json.Unmarshal(mpData, &mp); err != nil {
		return "", nil, err
	}
	if mp == nil {
		mp = map[string]interface{}{}
	}

	templateActionData, err := json.Marshal(templateAction)
	if err != nil {
		return "", nil, err
	}
	if err := json.Unmarshal(templateActionData, &templateSpec); err != nil {
		return "", nil, err
	}

	// TODO update after DiceHub refactor finish
	str, list, err := pipelineyml.DoRenderTemplateWithFormat(mp, templateSpec, alias, apistructs.TemplateVersion(templateVersion))
	if err != nil {
		return "", nil, err
	}
	listData, err := json.Marshal(list)
	if err != nil {
		return "", nil, err
	}
	if err := json.Unmarshal(listData, &respList); err != nil {
		return "", nil, err
	}

	return str, respList, nil
}
