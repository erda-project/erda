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

package handler_model

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"
)

func (h *ModelHandler) pagingViaDB(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	modelIDs, err := h.DAO.ModelClient().GetClientModelIDs(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}
	req.Ids = modelIDs
	if !req.ClientOnly {
		req.ClientId = ""
	}
	return h.DAO.ModelClient().Paging(ctx, req)
}

func (h *ModelHandler) pagingViaCache(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	allClientModels, err := cachehelpers.ListAllClientModels(ctx, req.ClientId, nil)
	if err != nil {
		return nil, err
	}
	var resultModels []*pb.Model
	reqIDsMap := make(map[string]struct{})
	for _, id := range req.Ids {
		reqIDsMap[id] = struct{}{}
	}
	for _, cm := range allClientModels {
		m := cm.Model
		// client only
		if req.ClientOnly {
			if m.ClientId == "" {
				continue
			}
		}
		// name like
		if req.Name != "" {
			if !strings.Contains(strings.ToLower(m.Name), strings.ToLower(req.Name)) {
				continue
			}
		}
		// name full
		if req.NameFull != "" {
			if !strings.EqualFold(m.Name, req.NameFull) {
				continue
			}
		}
		// ids
		if len(req.Ids) > 0 {
			if _, ok := reqIDsMap[m.Id]; !ok {
				continue
			}
		}
		// template id
		if req.TemplateId != "" {
			if m.TemplateId != req.TemplateId {
				continue
			}
		}
		// type
		if req.Type.Enum().Number() > 0 {
			if m.Type != req.Type {
				continue
			}
		}
		// provider id
		if req.ProviderId != "" {
			if m.ProviderId != req.ProviderId {
				continue
			}
		}
		// publisher
		if req.Publisher != "" {
			if m.Publisher != req.Publisher {
				continue
			}
		}
		// isEnabled
		if req.IsEnabled != nil {
			if m.IsEnabled == nil || *m.IsEnabled != *req.IsEnabled {
				continue
			}
		}
		resultModels = append(resultModels, m)
	}
	// sort by OrderBys
	if len(req.OrderBys) == 0 {
		req.OrderBys = append(req.OrderBys, "updated_at DESC")
	}
	sortModels(resultModels, req.OrderBys)

	// pagination
	total := int64(len(resultModels))
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	pageNum := int(req.PageNum)
	if pageNum <= 0 {
		pageNum = 1
	}
	start := (pageNum - 1) * pageSize
	if start > len(resultModels) {
		start = len(resultModels)
	}
	end := start + pageSize
	if end > len(resultModels) {
		end = len(resultModels)
	}
	pageList := resultModels[start:end]

	return &pb.ModelPagingResponse{
		Total: total,
		List:  pageList,
	}, nil
}

type orderBy struct {
	Field string
	Desc  bool
}

func parseOrderBys(orderBys []string) []orderBy {
	if len(orderBys) == 0 {
		orderBys = []string{"updated_at DESC"}
	}
	var res []orderBy
	for _, ob := range orderBys {
		parts := strings.Fields(ob)
		if len(parts) == 0 {
			continue
		}
		field := strings.ToLower(parts[0])
		desc := false
		if len(parts) > 1 && strings.EqualFold(parts[1], "desc") {
			desc = true
		}
		res = append(res, orderBy{
			Field: field,
			Desc:  desc,
		})
	}
	return res
}

func getUpdatedAt(m *pb.Model) time.Time {
	if m.UpdatedAt == nil {
		return time.Time{}
	}
	return m.UpdatedAt.AsTime()
}

func sortModels(models []*pb.Model, orderBys []string) {
	orders := parseOrderBys(orderBys)
	if len(orders) == 0 || len(models) == 0 {
		return
	}
	sort.SliceStable(models, func(i, j int) bool {
		a, b := models[i], models[j]
		for _, ob := range orders {
			switch ob.Field {
			case "updated_at":
				t1, t2 := getUpdatedAt(a), getUpdatedAt(b)
				if t1.Equal(t2) {
					continue
				}
				if ob.Desc {
					return t1.After(t2)
				}
				return t1.Before(t2)
			case "name":
				n1 := strings.ToLower(a.Name)
				n2 := strings.ToLower(b.Name)
				if n1 == n2 {
					continue
				}
				if ob.Desc {
					return n1 > n2
				}
				return n1 < n2
			default:
				// unsupported field, skip to next
				continue
			}
		}
		// all compared fields are equal, keep original order
		return false
	})
}
