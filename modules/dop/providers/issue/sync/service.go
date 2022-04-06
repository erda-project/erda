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

package sync

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issue"
	"github.com/erda-project/erda/pkg/common/apis"
)

type IssueSyncService struct {
	logger logs.Logger

	db    *dao.DBClient
	issue *issue.Issue
}

func (s *IssueSyncService) WithIssue(issue *issue.Issue) {
	s.issue = issue
}

func (s *IssueSyncService) IssueSync(ctx context.Context, req *pb.IssueSyncRequest) (*pb.IssueSyncResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}
	if req.Id == 0 {
		return nil, nil
	}
	issue, err := s.issue.GetIssue(apistructs.IssueGetRequest{ID: uint64(req.Id), IdentityInfo: apistructs.IdentityInfo{UserID: userID}})
	if err != nil {
		return nil, err
	}
	for _, i := range req.UpdateFields {
		// TODO: define issue field as enum and add more available issue fields for sync
		switch i.Field {
		case "labels":
			relatingIssueIDs, err := s.db.GetRelatingIssues(uint64(req.Id), []string{apistructs.IssueRelationInclusion})
			if err != nil {
				return nil, err
			}
			if err := s.issue.SyncLabels(i.Value, relatingIssueIDs); err != nil {
				return nil, err
			}
		case "iterationID":
			iterationID := int64(i.Value.Content.GetNumberValue())
			if err := s.issue.SyncIssueChildrenIteration(issue, iterationID); err != nil {
				return nil, err
			}
		}
	}

	return &pb.IssueSyncResponse{}, err
}
