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

package actionmgr

import (
	"context"
	"fmt"

	"sigs.k8s.io/yaml"

	extensionpb "github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/action/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/extension"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/limit_sync_group"
)

type actionService struct {
	p            *provider
	edgeRegister edgepipeline_register.Interface
	clusterInfo  clusterinfo.Interface
	extensionSvc extension.Interface
	bdl          *bundle.Bundle
}

func (s *actionService) CheckInternalClient(ctx context.Context) error {
	if apis.GetInternalClient(ctx) == "" {
		return fmt.Errorf("not internal client req")
	}
	return nil
}

func (s *actionService) List(ctx context.Context, req *pb.PipelineActionListRequest) (*extensionpb.QueryExtensionsResponse, error) {
	result, err := s.extensionSvc.QueryExtensionList(req.All, apistructs.SpecActionType.String(), req.Labels)
	if err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	locale := s.bdl.GetLocale(apis.GetLang(ctx))
	data, err := s.extensionSvc.MenuExtWithLocale(result, locale, req.All)
	if err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	var newResult []*extensionpb.Extension
	for _, menu := range data {
		for _, value := range menu {
			for _, extension := range value.Items {
				newResult = append(newResult, extension)
			}
		}
	}

	resp, err := s.extensionSvc.ToProtoValue(newResult)
	if err != nil {
		s.p.Log.Errorf("fail transform interface to any type")
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}
	return &extensionpb.QueryExtensionsResponse{Data: resp}, nil
}

func (s *actionService) Save(ctx context.Context, req *pb.PipelineActionSaveRequest) (*extensionpb.ExtensionVersionCreateResponse, error) {
	if err := s.CheckInternalClient(ctx); err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	if req.Spec == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("spec yml was empty")
	}
	if req.Dice == "" {
		return nil, apierrors.ErrSavePipelineAction.InvalidParameter("dice yml was empty")
	}

	extReq, err := transfer2ExtensionReq(req)
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}

	res, err := s.extensionSvc.CreateExtensionVersionByRequest(extReq)
	if err != nil {
		return nil, apierrors.ErrSavePipelineAction.InternalError(err)
	}
	s.syncActionToEdge(func(bdl *bundle.Bundle) error {
		_, err := bdl.SavePipelineAction(req)
		if err != nil {
			return err
		}
		return nil
	})

	return res, nil
}

func (s *actionService) syncActionToEdge(do func(bdl *bundle.Bundle) error) error {
	if s.edgeRegister.IsEdge() {
		return nil
	}

	edgeClusters, err := s.clusterInfo.ListEdgeClusterInfos()
	if err != nil {
		return err
	}

	wait := limit_sync_group.NewWorker(5)
	for index := range edgeClusters {
		wait.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
			index := i[0].(int)
			edgeCluster := edgeClusters[index]

			bdl, err := s.edgeRegister.GetEdgeBundleByClusterName(edgeCluster.Name)
			if err != nil {
				return err
			}
			return do(bdl)
		}, index)
	}

	return wait.Do().Error()
}

func transfer2ExtensionReq(req *pb.PipelineActionSaveRequest) (*extensionpb.ExtensionVersionCreateRequest, error) {
	var specInfo apistructs.Spec
	err := yaml.Unmarshal([]byte(req.Spec), &specInfo)
	if err != nil {
		return nil, err
	}

	extReq := &extensionpb.ExtensionVersionCreateRequest{}
	extReq.Name = specInfo.Name
	extReq.Version = specInfo.Version
	extReq.DiceYml = req.Dice
	extReq.SpecYml = req.Spec
	extReq.Readme = req.Readme
	extReq.Public = specInfo.Public
	extReq.IsDefault = specInfo.IsDefault
	extReq.ForceUpdate = true

	return extReq, nil
}

func (s *actionService) Delete(ctx context.Context, req *pb.PipelineActionDeleteRequest) (*pb.PipelineActionDeleteResponse, error) {
	if err := s.CheckInternalClient(ctx); err != nil {
		return nil, apierrors.ErrListPipelineAction.InternalError(err)
	}

	if req.Name == "" {
		return nil, apierrors.ErrDeletePipelineAction.InvalidParameter("name was empty")
	}
	if req.Version == "" {
		return nil, apierrors.ErrDeletePipelineAction.InvalidParameter("version was empty")
	}
	err := s.extensionSvc.DeleteExtensionVersion(req.Name, req.Version)
	if err != nil {
		return nil, apierrors.ErrDeletePipelineAction.InternalError(err)
	}
	s.syncActionToEdge(func(bdl *bundle.Bundle) error {
		return bdl.DeletePipelineAction(req)
	})

	return &pb.PipelineActionDeleteResponse{}, nil
}
