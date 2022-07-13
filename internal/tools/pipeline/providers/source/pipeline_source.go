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

package source

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/source/db"
)

type pipelineSource struct {
	dbClient *db.Client
}

func (p pipelineSource) Create(ctx context.Context, request *pb.PipelineSourceCreateRequest) (*pb.PipelineSourceCreateResponse, error) {
	unique := &db.PipelineSourceUnique{
		SourceType: request.SourceType,
		Remote:     request.Remote,
		Ref:        request.Ref,
		Path:       request.Path,
		Name:       request.Name,
	}

	var (
		source *db.PipelineSource
		err    error
	)

	sources, err := p.dbClient.GetPipelineSourceByUnique(unique)
	if err != nil {
		return nil, err
	}
	if len(sources) > 1 {
		return nil, fmt.Errorf("the pipeline source is not unique")
	}

	if len(sources) == 1 {
		return &pb.PipelineSourceCreateResponse{PipelineSource: sources[0].Convert()}, nil
	}

	source = &db.PipelineSource{
		SourceType:  request.SourceType,
		Remote:      request.Remote,
		Ref:         request.Ref,
		Path:        request.Path,
		Name:        request.Name,
		PipelineYml: request.PipelineYml,
		VersionLock: request.VersionLock,
	}
	if err = p.dbClient.CreatePipelineSource(source); err != nil {
		return nil, err
	}
	return &pb.PipelineSourceCreateResponse{PipelineSource: source.Convert()}, nil
}

func (p pipelineSource) Update(ctx context.Context, request *pb.PipelineSourceUpdateRequest) (*pb.PipelineSourceUpdateResponse, error) {
	source, err := p.dbClient.GetPipelineSource(request.PipelineSourceID)
	if err != nil {
		return nil, err
	}

	if request.PipelineYml != "" {
		source.PipelineYml = request.PipelineYml
	}
	err = p.dbClient.UpdatePipelineSource(request.PipelineSourceID, source)
	if err != nil {
		return nil, err
	}
	return &pb.PipelineSourceUpdateResponse{
		PipelineSource: source.Convert(),
	}, nil
}

func (p pipelineSource) Delete(ctx context.Context, request *pb.PipelineSourceDeleteRequest) (*pb.PipelineSourceDeleteResponse, error) {
	source, err := p.dbClient.GetPipelineSource(request.PipelineSourceID)
	if err != nil {
		return nil, err
	}
	source.SoftDeletedAt = uint64(time.Now().UnixNano() / 1e6)

	return &pb.PipelineSourceDeleteResponse{}, p.dbClient.DeletePipelineSource(request.PipelineSourceID, source)
}

func (p pipelineSource) DeleteByRemote(ctx context.Context, request *pb.PipelineSourceDeleteByRemoteRequest) (*pb.PipelineSourceDeleteResponse, error) {
	if request.Remote == "" {
		return nil, fmt.Errorf("the remote is empty")
	}

	sources, err := p.dbClient.ListPipelineSourceByRemote(request.Remote)
	if err != nil {
		return nil, err
	}
	session := p.dbClient.NewSession()
	defer session.Close()

	if err = session.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if rbErr := session.Rollback(); rbErr != nil {
				logrus.Errorf("failed to rollback when delete by remote, remote: %s, rollbackErr: %v",
					request.Remote, rbErr)
			}
			return
		}
		if cmErr := session.Commit(); cmErr != nil {
			logrus.Errorf("failed to commit when delete by remote, remote: %s, rollbackErr: %v",
				request.Remote, cmErr)
		}
	}()

	for _, v := range sources {
		v.SoftDeletedAt = uint64(time.Now().UnixNano() / 1e6)
		err = p.dbClient.DeletePipelineSource(v.ID, &v, mysqlxorm.WithSession(session))
		if err != nil {
			return nil, err
		}
	}

	return &pb.PipelineSourceDeleteResponse{}, nil
}

func (p pipelineSource) Get(ctx context.Context, request *pb.PipelineSourceGetRequest) (*pb.PipelineSourceGetResponse, error) {
	source, err := p.dbClient.GetPipelineSource(request.GetPipelineSourceID())
	if err != nil {
		return nil, err
	}
	return &pb.PipelineSourceGetResponse{PipelineSource: source.Convert()}, nil
}

func (p pipelineSource) List(ctx context.Context, request *pb.PipelineSourceListRequest) (*pb.PipelineSourceListResponse, error) {
	unique := &db.PipelineSourceUnique{
		SourceType: request.SourceType,
		Remote:     request.Remote,
		Ref:        request.Ref,
		Path:       request.Path,
		Name:       request.Name,
		IDList:     request.IdList,
	}

	var sources []db.PipelineSource
	var err error
	if request.IdList != nil {
		sources, err = p.dbClient.ListPipelineSource(request.IdList)
	} else {
		sources, err = p.dbClient.GetPipelineSourceByUnique(unique)
	}
	if err != nil {
		return nil, err
	}

	data := make([]*pb.PipelineSource, 0, len(sources))
	for _, v := range sources {
		data = append(data, v.Convert())
	}

	return &pb.PipelineSourceListResponse{Data: data}, nil
}
