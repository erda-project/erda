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

package report

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/pipeline/report/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/strutil"
)

type reportService struct {
	p        *provider
	dbClient *dbclient.Client
}

func (s *reportService) QueryPipelineReportSet(ctx context.Context, req *pb.PipelineReportSetQueryRequest) (*pb.PipelineReportSetQueryResponse, error) {
	dbReports, err := s.dbClient.BatchListPipelineReportsByPipelineID([]uint64{req.PipelineID}, req.Types)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineReportSet.InternalError(err)
	}
	var reports []*pb.PipelineReport
	for _, v := range dbReports[req.PipelineID] {
		pbReport, err := v.ConvertToPB()
		if err != nil {
			return nil, err
		}
		reports = append(reports, pbReport)
	}
	return &pb.PipelineReportSetQueryResponse{
		Data: &pb.PipelineReportSet{
			PipelineID: req.PipelineID,
			Reports:    reports,
		},
	}, nil
}
func (s *reportService) PagingPipelineReportSet(ctx context.Context, req *pb.PipelineReportSetPagingRequest) (*pb.PipelineReportSetPagingResponse, error) {
	sets, total, err := s.dbClient.PagingPipelineReportSets(req)
	if err != nil {
		return nil, apierrors.ErrQueryPipelineReportSet.InternalError(err)
	}
	return &pb.PipelineReportSetPagingResponse{
		Total: int64(total),
		Data:  sets,
	}, nil
}

func (s *reportService) Create(req *pb.PipelineReportCreateRequest) (*pb.PipelineReport, error) {
	if err := s.ValidatePipelineReportReq(req); err != nil {
		return nil, apierrors.ErrCreatePipelineReport.InvalidParameter(err)
	}
	p, exist, err := s.dbClient.GetPipelineBase(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineReport.InvalidParameter(fmt.Errorf("failed to find pipeline, err: %v", err))
	}
	if !exist {
		return nil, apierrors.ErrCreatePipelineReport.InvalidParameter(fmt.Errorf("pipeline not exist"))
	}
	dbReport := spec.PipelineReport{
		PipelineID: req.PipelineID,
		Type:       apistructs.PipelineReportType(req.Type),
		Meta:       req.Meta.AsMap(),
	}
	if req.IdentityInfo != nil {
		dbReport.CreatorID = req.IdentityInfo.UserID
		dbReport.UpdaterID = req.IdentityInfo.UserID
	}
	if err := s.dbClient.CreatePipelineReport(&dbReport); err != nil {
		return nil, apierrors.ErrCreatePipelineReport.InternalError(fmt.Errorf("failed to create report in database, err: %v", err))
	}
	reportLabelKey, reportLabelValue := s.dbClient.MakePipelineReportTypeLabelKey(apistructs.PipelineReportType(req.Type))
	if err := s.dbClient.CreatePipelineLabels(&spec.Pipeline{
		PipelineBase: spec.PipelineBase{ID: p.ID, PipelineSource: p.PipelineSource, PipelineYmlName: p.PipelineYmlName},
		Labels:       map[string]string{reportLabelKey: reportLabelValue},
	}); err != nil {
		return nil, apierrors.ErrCreatePipelineReport.InternalError(fmt.Errorf("failed to create related pipeline labels, err: %v", err))
	}
	report, err := dbReport.ConvertToPB()
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (s *reportService) MakePBMeta(meta interface{}) (*structpb.Struct, error) {
	var reqMeta map[string]interface{}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &reqMeta); err != nil {
		return nil, err
	}
	pbMeta, err := structpb.NewStruct(reqMeta)
	if err != nil {
		return nil, err
	}
	return pbMeta, nil
}

func (s *reportService) ValidatePipelineReportReq(req *pb.PipelineReportCreateRequest) error {
	if req.PipelineID == 0 {
		return fmt.Errorf("missing pipelineID")
	}
	if err := strutil.Validate(string(req.Type), strutil.MinLenValidator(1), strutil.MaxLenValidator(32)); err != nil {
		return fmt.Errorf("invalid type: %v", err)
	}
	return nil
}
