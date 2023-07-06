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

package handlers

import (
	"context"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"regexp"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

var (
	_ pb.CredentialsServer = (*CredentialsHandler)(nil)
)

type CredentialsHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (h *CredentialsHandler) Q() *gorm.DB {
	return h.Dao.Q()
}

func (h *CredentialsHandler) CreateCredential(ctx context.Context, credential *pb.Credential) (*pb.Credential, error) {
	AdjustCredential(credential)
	if err := CheckCredential(credential); err != nil {
		return nil, err
	}
	if err := h.Q().First(new(models.AIProxyCredentials), map[string]any{
		"platform": credential.GetPlatform(),
		"name":     credential.GetName(),
	}).Error; err == nil {
		return nil, errors.Errorf("the credential % on the platform %s already exists", credential.GetName(), credential.GetPlatform())
	}
	var model = models.NewCredential(credential)
	if err := h.Dao.Create(&model).Error; err != nil {
		return nil, errors.Wrap(err, "failed to create credential")
	}
	return model.ToProtobuf(), nil
}

func (h *CredentialsHandler) DeleteCredential(_ context.Context, req *pb.DeleteCredentialReq) (*common.VoidResponse, error) {
	return new(common.VoidResponse), h.Q().Delete(new(models.AIProxyCredentials), map[string]any{"access_key_id": req.GetAccessKeyId()}).Error
}

func (h *CredentialsHandler) UpdateCredential(_ context.Context, credential *pb.Credential) (*pb.Credential, error) {
	if err := CheckCredential(credential); err != nil {
		return nil, err
	}
	var model models.AIProxyCredentials
	where := map[string]any{"access_key_id": credential.GetAccessKeyId()}
	if err := h.Q().First(&model, where).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find the credential")
	}
	model.SecretKeyId = credential.GetSecretKeyId()
	model.Name = credential.GetName()
	model.Platform = credential.GetPlatform()
	model.Description = credential.GetDescription()
	model.Enabled = credential.GetEnabled()
	var updates = map[string]any{
		"secret_key_id":     model.SecretKeyId,
		"name":              model.Name,
		"platform":          model.Platform,
		"description":       model.Description,
		"enabled":           model.Enabled,
		"provider_name":     model.ProviderName,
		"provider_instance": model.ProviderInstance,
	}
	if err := h.Dao.Model(&model).Where(where).Updates(updates).Error; err != nil {
		return nil, errors.Wrap(err, "failed to update credential")
	}
	return model.ToProtobuf(), nil
}

func (h *CredentialsHandler) ListCredentials(_ context.Context, credential *pb.Credential) (*pb.ListCredentialsRespData, error) {
	// todo: condition and count

	var credentials []*models.AIProxyCredentials
	if err := h.Dao.Find(&credentials).Error; err != nil {
		return nil, errors.Wrap(err, "failed to find credentials")
	}
	var total = len(credentials)
	var list = make([]*pb.Credential, total)
	for i := 0; i < total; i++ {
		list[i] = credentials[i].ToProtobuf()
	}
	return &pb.ListCredentialsRespData{
		Total: uint32(total),
		List:  list,
	}, nil
}

func (h *CredentialsHandler) GetCredential(_ context.Context, req *pb.GetCredentialReq) (*pb.Credential, error) {
	var model models.AIProxyCredentials
	if err := h.Q().First(&model, map[string]any{"access_key_id": req.GetAccessKeyId()}).Error; err != nil {
		return nil, errors.Wrap(err, "failed to get credential")
	}
	return model.ToProtobuf(), nil
}

func AdjustCredential(credential *pb.Credential) {
	if credential.GetAccessKeyId() == "" {
		credential.AccessKeyId = uuid.New().String()
	}
	if credential.GetSecretKeyId() == "" {
		credential.SecretKeyId = uuid.New().String()
	}
}

func CheckCredential(credential *pb.Credential) error {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`
	re := regexp.MustCompile(pattern)
	if ok := re.MatchString(credential.GetAccessKeyId()); !ok {
		return errors.New("invalid accessKeyId, it must be an UUID")
	}
	if ok := re.MatchString(credential.GetSecretKeyId()); !ok {
		return errors.New("invalid secretKeyId, it must be an UUID")
	}
	if credential.GetName() == "" {
		return errors.New("credential must have a name")
	}
	if credential.GetPlatform() == "" {
		return errors.New("credential must on a platform")
	}
	return nil
}
