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

package home

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/dop/projecthome/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type projectHomeService struct {
	logger logs.Logger

	db  *dao.DBClient
	bdl *bundle.Bundle
}

func (s *projectHomeService) GetProjectHome(ctx context.Context, req *pb.GetProjectHomeRequest) (*pb.GetProjectHomeResponse, error) {
	projectID, err := strconv.ParseUint(req.ProjectID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid projectID: %s, err: %v", req.ProjectID, err)
	}
	_, err = s.bdl.GetProjectWithSetter(projectID, httpclient.SetHeaders(
		http.Header{
			httputil.InternalHeader: []string{""},
			httputil.OrgHeader:      []string{apis.GetOrgID(ctx)},
			httputil.UserHeader:     []string{apis.GetUserID(ctx)},
		},
	))
	if err != nil {
		return nil, err
	}
	home, err := s.db.GetProjectHome(req.ProjectID)
	if err != nil {
		return nil, err
	}
	if home == nil {
		return &pb.GetProjectHomeResponse{Data: &pb.ProjectHome{}}, nil
	}
	var links []*pb.Link
	if err := json.Unmarshal([]byte(home.Links), &links); err != nil {
		return nil, err
	}
	return &pb.GetProjectHomeResponse{Data: &pb.ProjectHome{Readme: home.Readme, Links: links}}, nil
}

func (s *projectHomeService) CreateOrUpdateProjectHome(ctx context.Context, req *pb.CreateOrUpdateProjectHomeRequest) (*pb.CreateOrUpdateProjectHomeResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, fmt.Errorf("not login")
	}
	linkStr, err := json.Marshal(req.Links)
	if err != nil {
		return nil, err
	}
	links := string(linkStr)
	return &pb.CreateOrUpdateProjectHomeResponse{}, s.db.CreateOrUpdateProjectHome(req.ProjectID, req.Readme, links, userID)
}
