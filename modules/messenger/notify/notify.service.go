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

package notify

import (
	"context"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/core/messenger/notify/pb"
	"github.com/erda-project/erda/modules/messenger/notify/db"
	"github.com/erda-project/erda/modules/messenger/notify/model"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type notifyService struct {
	DB *db.DB
	L  logs.Logger
}

func (n notifyService) CreateNotifyHistory(ctx context.Context, request *pb.CreateNotifyHistoryRequest) (*pb.CreateNotifyHistoryResponse, error) {
	result := &pb.CreateNotifyHistoryResponse{}
	var historyId int64
	var err error
	if request.AlertID != 0 {
		historyId, err = n.CreateHistoryAndIndex(request)
		if err != nil {
			return result, errors.NewInternalServerError(err)
		}
	} else {
		history, err := n.DB.NotifyHistoryDB.CreateNotifyHistory(request)
		if err != nil {
			return result, errors.NewInternalServerError(err)
		}
		historyId = history.ID
	}
	result.Data = historyId
	return result, nil
}

func (n notifyService) CreateHistoryAndIndex(request *pb.CreateNotifyHistoryRequest) (historyId int64, err error) {
	tx := n.DB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			n.L.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	history, err := tx.NotifyHistoryDB.CreateNotifyHistory(request)
	if err != nil {
		return 0, err
	}
	alertNotifyIndex := &db.AlertNotifyIndex{
		NotifyID:   history.ID,
		NotifyName: request.NotifyName,
		Status:     request.Status,
		Channel:    request.Channel,
		AlertID:    request.AlertID,
		CreatedAt:  time.Now(),
		SendTime:   history.CreatedAt,
	}
	_, err = tx.AlertNotifyIndexDB.CreateAlertNotifyIndex(alertNotifyIndex)
	if err != nil {
		return 0, err
	}
	return history.ID, nil
}

func (n notifyService) QueryNotifyHistories(ctx context.Context, request *pb.QueryNotifyHistoriesRequest) (*pb.QueryNotifyHistoriesResponse, error) {
	orgIdStr := apis.GetOrgID(ctx)
	orgId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		return nil, errors.NewInternalServerError(err)
	}
	request.OrgID = int64(orgId)
	list, count, err := n.DB.NotifyHistoryDB.QueryNotifyHistories(request)
	if err != nil {
		return &pb.QueryNotifyHistoriesResponse{}, nil
	}
	result := &pb.QueryNotifyHistoriesResponse{
		Data: &pb.QueryNotifyHistoryData{
			List: []*pb.NotifyHistory{},
		},
	}
	for _, notifyHistory := range list {
		notify, err := notifyHistory.ToApiData()
		if err != nil {
			return result, nil
		}
		result.Data.List = append(result.Data.List, notify)
	}
	result.Data.Total = count
	return result, nil
}

func (n notifyService) GetNotifyStatus(ctx context.Context, request *pb.GetNotifyStatusRequest) (*pb.GetNotifyStatusResponse, error) {
	orgIdStr := apis.GetOrgID(ctx)
	result := &pb.GetNotifyStatusResponse{
		Data: make(map[string]int64),
	}
	orgId, err := strconv.Atoi(orgIdStr)
	if err != nil {
		return result, errors.NewInternalServerError(err)
	}
	filterStatus := &model.FilterStatusRequest{
		OrgId:     orgId,
		ScopeType: request.ScopeType,
		ScopeId:   request.ScopeId,
		StartTime: request.StartTime,
		EndTime:   request.EndTime,
	}
	filterStatusResult, err := n.DB.NotifyHistoryDB.FilterStatus(filterStatus)
	if err != nil {
		return result, errors.NewInternalServerError(err)
	}
	for _, v := range filterStatusResult {
		result.Data[v.Status] = v.Count
	}
	return result, nil
}
