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

package models

import (
	"strconv"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

func (this *AIProxySessions) ToProtobuf() *pb.Session {
	temperature, _ := strconv.ParseFloat(this.Temperature, 10)
	return &pb.Session{
		Id:            this.ID.String,
		UserId:        this.UserID,
		Name:          this.Name,
		Topic:         this.Topic,
		ContextLength: uint32(this.ContextLength),
		IsArchived:    this.IsArchived,
		Source:        this.Source,
		ResetAt:       timestamppb.New(this.ResetAt),
		Model:         this.Model,
		Temperature:   temperature,
		CreatedAt:     timestamppb.New(this.CreatedAt),
		UpdatedAt:     timestamppb.New(this.UpdatedAt),
	}
}

func (list AIProxySessionsList) ToProtobuf() (result []*pb.Session) {
	for _, item := range list {
		result = append(result, item.ToProtobuf())
	}
	return result
}
