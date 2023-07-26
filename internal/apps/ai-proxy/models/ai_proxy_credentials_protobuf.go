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
	"time"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
)

func NewCredential(credential *pb.Credential) *AIProxyCredentials {
	var model AIProxyCredentials
	model.AccessKeyID = credential.GetAccessKeyId()
	model.SecretKeyID = credential.GetSecretKeyId()
	model.Name = credential.GetName()
	model.Platform = credential.GetPlatform()
	model.Description = credential.GetDescription()
	model.Enabled = credential.GetEnabled()
	model.ExpiredAt = time.Date(2099, time.January, 1, 0, 0, 0, 0, time.Local)
	model.ProviderName = credential.GetProviderName()
	model.ProviderInstanceID = credential.GetProviderInstanceId()
	return &model
}

func (model AIProxyCredentials) ToProtobuf() *pb.Credential {
	return &pb.Credential{
		AccessKeyId:        model.AccessKeyID,
		SecretKeyId:        model.SecretKeyID,
		Name:               model.Name,
		Platform:           model.Platform,
		Description:        model.Description,
		Enabled:            model.Enabled,
		ProviderName:       model.ProviderName,
		ProviderInstanceId: model.ProviderInstanceID,
	}
}
