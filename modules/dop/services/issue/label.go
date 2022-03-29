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

package issue

import (
	"github.com/erda-project/erda-proto-go/dop/issue/sync/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
)

func (svc *Issue) UpdateLabels(id, projectID uint64, labelNames []string) (err error) {
	if err = svc.db.DeleteLabelRelations(apistructs.LabelTypeIssue, id, nil); err != nil {
		return
	}
	labels, err := svc.bdl.ListLabelByNameAndProjectID(projectID, labelNames)
	if err != nil {
		return
	}
	labelRelations := make([]dao.LabelRelation, 0, len(labels))
	for _, v := range labels {
		labelRelations = append(labelRelations, dao.LabelRelation{
			LabelID: uint64(v.ID),
			RefType: apistructs.LabelTypeIssue,
			RefID:   id,
		})
	}
	return svc.db.BatchCreateLabelRelations(labelRelations)
}

func (s *Issue) SyncLabels(req *pb.IssueSyncRequest, issueIDs []uint64) error {
	if req.Addition == nil || req.Deletion == nil {
		return nil
	}
	labelsAdd := req.Addition.Labels
	labelsDelete := req.Deletion.Labels
	for _, id := range issueIDs {
		issue, err := s.GetIssue(apistructs.IssueGetRequest{ID: id, IdentityInfo: apistructs.IdentityInfo{InternalClient: apistructs.SystemOperator}})
		if err != nil {
			return err
		}
		currentLabelMap := make(map[int64]bool)
		for _, v := range issue.LabelDetails {
			currentLabelMap[v.ID] = true
		}
		labelRelations := make([]dao.LabelRelation, 0, len(labelsAdd))
		for _, v := range labelsAdd {
			if _, ok := currentLabelMap[int64(v)]; !ok {
				labelRelations = append(labelRelations, dao.LabelRelation{
					LabelID: v,
					RefType: apistructs.LabelTypeIssue,
					RefID:   id,
				})
			}
		}
		if len(labelRelations) > 0 {
			if err := s.db.BatchCreateLabelRelations(labelRelations); err != nil {
				return err
			}
		}

		labelIDs := make([]uint64, 0, len(labelsDelete))
		for _, v := range labelsDelete {
			if _, ok := currentLabelMap[int64(v)]; ok {
				labelIDs = append(labelIDs, v)
			}
		}
		if len(labelIDs) > 0 {
			if err := s.db.DeleteLabelRelations(apistructs.LabelTypeIssue, id, labelIDs); err != nil {
				return err
			}
		}

		if err = s.stream.CreateIssueStreamBySystem(id, map[string][]interface{}{
			"label": {"", "", apistructs.ParentLabelsChanged},
		}); err != nil {
			return err
		}
	}
	return nil
}
