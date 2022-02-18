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

package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/dop/issue/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/pkg/common/apis"
)

type commentIssueStreamService struct {
	logger logs.Logger

	db *dao.DBClient
}

func (s *commentIssueStreamService) BatchCreateIssueStream(ctx context.Context, req *pb.CommentIssueStreamBatchCreateRequest) (*pb.CommentIssueStreamBatchCreateResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}

	if len(req.IssueStreams) == 0 {
		return &pb.CommentIssueStreamBatchCreateResponse{}, nil
	}

	issueStreams := make([]dao.IssueStream, 0, len(req.IssueStreams))
	issueAppRels := make([]dao.IssueAppRelation, 0, len(req.IssueStreams))
	for _, i := range req.IssueStreams {
		if i.Type == "" {
			i.Type = string(apistructs.ISTComment)
		}

		var istParam apistructs.ISTParam
		if i.Type == string(apistructs.ISTComment) {
			istParam.Comment = i.Content
			istParam.CommentTime = time.Now().Format("2006-01-02 15:04:05")
		} else {
			istParam.MRInfo = apistructs.MRCommentInfo{
				AppID:   i.MrInfo.AppID,
				MRID:    i.MrInfo.MrID,
				MRTitle: i.MrInfo.MrTitle,
			}
			issueAppRel := dao.IssueAppRelation{
				IssueID: i.IssueID,
				AppID:   i.MrInfo.AppID,
				MRID:    i.MrInfo.MrID,
			}
			issueAppRels = append(issueAppRels, issueAppRel)
		}

		is := dao.IssueStream{
			IssueID:      i.IssueID,
			Operator:     i.UserID,
			StreamType:   apistructs.IssueStreamType(i.Type),
			StreamParams: istParam,
		}
		issueStreams = append(issueStreams, is)
	}

	if err := s.db.BatchCreateIssueStream(issueStreams); err != nil {
		return nil, err
	}

	if len(issueAppRels) > 0 {
		if err := s.db.BatchCreateIssueAppRelation(issueAppRels); err != nil {
			return nil, err
		}
	}
	return &pb.CommentIssueStreamBatchCreateResponse{}, nil
}
