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
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/session/pb"
)

func (this *AIProxySession) ToProtobuf() *pb.Session {
	return &pb.Session{
		Id:          this.ID.String,
		CreatedAt:   timestamppb.New(this.CreatedAt),
		UpdatedAt:   timestamppb.New(this.UpdatedAt),
		DeletedAt:   timestamppb.New(this.DeletedAt.Time),
		ClientId:    this.ClientID,
		PromptId:    this.PromptID,
		ModelId:     this.ModelID,
		Scene:       this.Scene,
		UserId:      this.UserID,
		Name:        this.Name,
		Topic:       this.Topic,
		NumOfCtxMsg: this.NumOfCtxMsg,
		IsArchived:  this.IsArchived,
		ResetAt:     timestamppb.New(this.ResetAt),
		Temperature: this.Temperature,
		Metadata:    this.Metadata.ToProtobuf(),
	}
}

func (list *AIProxySessionList) ToProtobuf() []*pb.Session {
	var ret []*pb.Session
	for _, session := range *list {
		ret = append(ret, session.ToProtobuf())
	}
	return ret
}
