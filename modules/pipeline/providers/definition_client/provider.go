// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package definition_client

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition_client/deftype"
)

type config struct {
}

// +provider
type Provider struct {
	Cfg                     *config
	ClientDefinitionService pb.DefinitionServiceServer `autowired:"erda.core.pipeline.definition.DefinitionService"`
}

type Processor interface {
	ProcessPipelineDefinition(ctx context.Context, req deftype.ClientDefinitionProcessRequest) (*deftype.ClientDefinitionProcessResponse, error)
}

func (p *Provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *Provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

func (p *Provider) ProcessPipelineDefinition(ctx context.Context, req deftype.ClientDefinitionProcessRequest) (*deftype.ClientDefinitionProcessResponse, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context was empty")
	}

	result, err := p.ClientDefinitionService.Version(ctx, &pb.PipelineDefinitionProcessVersionRequest{
		PipelineSource:  req.PipelineSource.String(),
		PipelineYmlName: req.PipelineYmlName,
	})
	if err != nil {
		return nil, err
	}
	var processReq = pb.PipelineDefinitionProcessRequest{
		PipelineYmlName: req.PipelineYmlName,
		PipelineSource:  req.PipelineSource.String(),
		IsDelete:        req.IsDelete,
		PipelineYml:     req.PipelineYml,
		VersionLock:     result.VersionLock,
	}

	if req.PipelineCreateRequest != nil {
		var value = structpb.Value{}
		data, err := json.Marshal(req.PipelineCreateRequest)
		if err != nil {
			return nil, err
		}
		err = value.UnmarshalJSON(data)
		if err != nil {
			return nil, err
		}
		processReq.PipelineCreateRequest = &value
	}

	if req.SnippetConfig != nil {
		var value = structpb.Value{}
		data, err := json.Marshal(req.SnippetConfig)
		if err != nil {
			return nil, err
		}
		err = value.UnmarshalJSON(data)
		if err != nil {
			return nil, err
		}
		processReq.SnippetConfig = &value
	}

	// process will check pipelineYml value Don't care if you delete
	if processReq.IsDelete == true {
		processReq.PipelineYml = "version: \"1.1\"\nstages: []"
	}

	processResult, err := p.ClientDefinitionService.Process(ctx, &processReq)
	if err != nil {
		return nil, err
	}

	return &deftype.ClientDefinitionProcessResponse{
		ID:              processResult.Id,
		PipelineSource:  apistructs.PipelineSource(processResult.PipelineSource),
		PipelineYmlName: processResult.PipelineYmlName,
		VersionLock:     processResult.VersionLock,
		PipelineYml:     processResult.PipelineYml,
	}, nil
}

func init() {
	servicehub.Register("erda.core.pipeline.definition-process-client", &servicehub.Spec{
		Services:             []string{"erda.core.pipeline.definition-process-client"},
		Types:                []reflect.Type{reflect.TypeOf(reflect.TypeOf((*Processor)(nil)).Elem())},
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &Provider{}
		},
	})
}
