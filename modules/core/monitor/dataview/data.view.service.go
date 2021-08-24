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

package dataview

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-proto-go/core/monitor/dataview/pb"
	"github.com/erda-project/erda/modules/core/monitor/dataview/db"
	"github.com/erda-project/erda/pkg/common/errors"
)

type dataViewService struct {
	p      *provider
	sys    *db.SystemViewDB
	custom *db.CustomViewDB
}

func (s *dataViewService) parseViewBlocks(view *pb.View, config, data string) *pb.View {
	if len(config) > 0 {
		err := json.Unmarshal([]byte(config), &view.Blocks)
		if err != nil {
			s.p.Log.Warn("invalid config in view (%s): %s", view.Id, err)
			return view
		}
		if len(data) > 0 {
			var values []struct {
				I          string          `json:"i"`
				StaticData *structpb.Value `json:"staticData"`
			}
			err = json.Unmarshal([]byte(data), &values)
			if err != nil {
				s.p.Log.Warn("invalid data in view (%s): %s", view.Id, err)
				return view
			}
			empty, _ := structpb.NewValue(map[string]interface{}{})
			for _, block := range view.Blocks {
				if block == nil || block.Chart == nil {
					continue
				}
				block.Chart.StaticData = empty
				for _, val := range values {
					if block.I == val.I {
						block.Chart.StaticData = val.StaticData
					}
				}
			}
		}
	}
	return view
}

func (s *dataViewService) ListSystemViews(ctx context.Context, req *pb.ListSystemViewsRequest) (*pb.ListSystemViewsResponse, error) {
	list, err := s.sys.ListByFields(map[string]interface{}{
		"Scope":   req.Scope,
		"ScopeID": req.ScopeID,
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	views := &pb.ViewList{}
	for _, item := range list {
		view := s.parseViewBlocks(&pb.View{
			Id:        item.ID,
			Scope:     item.Scope,
			ScopeID:   item.ScopeID,
			Version:   item.Version,
			Name:      item.Name,
			Desc:      item.Desc,
			CreatedAt: item.CreatedAt.UnixNano() / int64(time.Millisecond),
			UpdatedAt: item.UpdatedAt.UnixNano() / int64(time.Millisecond),
		}, item.ViewConfig, item.DataConfig)
		views.List = append(views.List, view)
	}
	views.Total = int64(len(views.List))
	return &pb.ListSystemViewsResponse{Data: views}, nil
}

func (s *dataViewService) GetSystemView(ctx context.Context, req *pb.GetSystemViewRequest) (*pb.GetSystemViewResponse, error) {
	data, err := s.sys.GetByFields(map[string]interface{}{
		"ID": req.Id,
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if data == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("view/%s", req.Id))
	}
	view := s.parseViewBlocks(&pb.View{
		Id:        data.ID,
		Scope:     data.Scope,
		ScopeID:   data.ScopeID,
		Version:   data.Version,
		Name:      data.Name,
		Desc:      data.Desc,
		CreatedAt: data.CreatedAt.UnixNano() / int64(time.Millisecond),
		UpdatedAt: data.UpdatedAt.UnixNano() / int64(time.Millisecond),
	}, data.ViewConfig, data.DataConfig)
	return &pb.GetSystemViewResponse{Data: view}, nil
}

func (s *dataViewService) ListCustomViews(ctx context.Context, req *pb.ListCustomViewsRequest) (*pb.ListCustomViewsResponse, error) {
	list, err := s.custom.ListByFields(map[string]interface{}{
		"Scope":   req.Scope,
		"ScopeID": req.ScopeID,
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	views := &pb.ViewList{}
	for _, item := range list {
		view := s.parseViewBlocks(&pb.View{
			Id:        item.ID,
			Scope:     item.Scope,
			ScopeID:   item.ScopeID,
			Version:   item.Version,
			Name:      item.Name,
			Desc:      item.Desc,
			CreatedAt: item.CreatedAt.UnixNano() / int64(time.Millisecond),
			UpdatedAt: item.UpdatedAt.UnixNano() / int64(time.Millisecond),
		}, item.ViewConfig, item.DataConfig)
		views.List = append(views.List, view)
	}
	views.Total = int64(len(views.List))
	return &pb.ListCustomViewsResponse{Data: views}, nil
}

func (s *dataViewService) GetCustomView(ctx context.Context, req *pb.GetCustomViewRequest) (*pb.GetCustomViewResponse, error) {
	data, err := s.custom.GetByFields(map[string]interface{}{
		"ID": req.Id,
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	if data == nil {
		return nil, errors.NewNotFoundError(fmt.Sprintf("view/%s", req.Id))
	}
	view := s.parseViewBlocks(&pb.View{
		Id:        data.ID,
		Scope:     data.Scope,
		ScopeID:   data.ScopeID,
		Version:   data.Version,
		Name:      data.Name,
		Desc:      data.Desc,
		CreatedAt: data.CreatedAt.UnixNano() / int64(time.Millisecond),
		UpdatedAt: data.UpdatedAt.UnixNano() / int64(time.Millisecond),
	}, data.ViewConfig, data.DataConfig)
	return &pb.GetCustomViewResponse{Data: view}, nil
}

func (s *dataViewService) CreateCustomView(ctx context.Context, req *pb.CreateCustomViewRequest) (*pb.CreateCustomViewResponse, error) {
	if req.Blocks == nil {
		req.Blocks = make([]*pb.Block, 0)
	}
	if req.Data == nil {
		req.Data = make([]*pb.DataItem, 0)
	}
	blocks, _ := json.Marshal(req.Blocks)
	data, _ := json.Marshal(req.Data)
	now := time.Now()
	model := &db.CustomView{
		ID:         req.Id,
		Name:       req.Name,
		Version:    req.Version,
		Desc:       req.Desc,
		Scope:      req.Scope,
		ScopeID:    req.ScopeID,
		ViewConfig: string(blocks),
		DataConfig: string(data),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if len(model.ID) <= 0 {
		model.ID = hex.EncodeToString(uuid.NewV4().Bytes())
	}
	err := s.custom.DB.Save(model).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.CreateCustomViewResponse{Data: s.parseViewBlocks(&pb.View{
		Id:        model.ID,
		Name:      model.Name,
		Version:   model.Version,
		Desc:      model.Desc,
		Scope:     model.Scope,
		ScopeID:   model.ScopeID,
		CreatedAt: model.CreatedAt.UnixNano() / int64(time.Millisecond),
		UpdatedAt: model.UpdatedAt.UnixNano() / int64(time.Millisecond),
	}, model.ViewConfig, model.DataConfig)}, nil
}

func (s *dataViewService) UpdateCustomView(ctx context.Context, req *pb.UpdateCustomViewRequest) (*pb.UpdateCustomViewResponse, error) {
	blocks, _ := json.Marshal(req.Blocks)
	data, _ := json.Marshal(req.Data)
	// TODO: support update some field
	err := s.custom.UpdateView(req.Id, map[string]interface{}{
		"Name":       req.Name,
		"Desc":       req.Desc,
		"ViewConfig": string(blocks),
		"DataConfig": string(data),
	})
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.UpdateCustomViewResponse{Data: true}, nil
}

func (s *dataViewService) DeleteCustomView(ctx context.Context, req *pb.DeleteCustomViewRequest) (*pb.DeleteCustomViewResponse, error) {
	err := s.custom.DB.Where("id=?", req.Id).Delete(&db.CustomView{}).Error
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return &pb.DeleteCustomViewResponse{Data: true}, nil
}
