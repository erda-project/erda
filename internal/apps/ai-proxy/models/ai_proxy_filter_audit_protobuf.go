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
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

func (audit *AIProxyFilterAudit) ToProtobufChatLog() *pb.ChatLog {
	return &pb.ChatLog{
		Id:         audit.ID.String,
		RequestAt:  timestamppb.New(audit.RequestAt),
		Prompt:     audit.Prompt,
		ResponseAt: timestamppb.New(audit.ResponseAt),
		Completion: audit.Completion,
	}
}

func (audit *AIProxyFilterAudit) SetAPIKeySha256(apiKey string) {
	audit.APIKeySHA256 = Sha256(apiKey)
}

func Sha256(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
