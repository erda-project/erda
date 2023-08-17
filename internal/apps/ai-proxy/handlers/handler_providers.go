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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	common "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/provider"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	_ pb.AIProviderServer = (*ProviderHandler)(nil)
)

type ProviderHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (p *ProviderHandler) Sync(_ context.Context, providers provider.Providers) error {
	for _, item := range providers {
		var prov = new(models.AIProxyProviders)
		var where = []models.Where{
			prov.FieldName().Equal(item.Name),
			prov.FieldInstanceID().Equal(item.InstanceId),
		}

		ok, err := prov.Retriever(p.Dao.Q()).Where(where...).Get()
		if err != nil {
			return err
		}
		prov.FromProtobuf(&pb.Provider{
			Name:        item.Name,
			InstanceId:  item.InstanceId,
			Host:        item.GetHost(),
			Scheme:      item.Scheme,
			Description: item.Description,
			DocSite:     item.DocSite,
			ApiKey:      item.GetAppKey(),
			Metadata:    strutil.TryGetJsonStr(item.Metadata),
		})
		if !ok {
			if err := prov.Creator(p.Dao.Q()).Create(); err != nil {
				return err
			}
			continue
		}
		if _, err := prov.Updater(p.Dao.Q()).Where(where...).Updates(
			prov.FieldName().Set(prov.Name),
			prov.FieldInstanceID().Set(prov.InstanceID),
			prov.FieldHost().Set(prov.Host),
			prov.FieldScheme().Set(prov.Scheme),
			prov.FieldDescription().Set(prov.Description),
			prov.FieldDocSite().Set(prov.DocSite),
			prov.FieldAesKey().Set(prov.AesKey),
			prov.FieldAPIKey().Set(prov.APIKey),
			prov.FieldMetadata().Set(prov.Metadata),
		); err != nil {
			return err
		}
	}
	return nil
}

func (p *ProviderHandler) CreateProvider(ctx context.Context, provider *pb.Provider) (*common.VoidResponse, error) {
	return new(common.VoidResponse), new(models.AIProxyProviders).FromProtobuf(provider).Creator(p.Dao.Q()).Create()
}

func (p *ProviderHandler) DeleteProvider(ctx context.Context, provider *pb.Provider) (*common.VoidResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (p *ProviderHandler) UpdateProvider(ctx context.Context, provider *pb.Provider) (*pb.Provider, error) {
	//TODO implement me
	panic("implement me")
}

func (p *ProviderHandler) ListProviders(ctx context.Context, provider *pb.Provider) (*pb.ListProvidersRespData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *ProviderHandler) GetProvider(ctx context.Context, provider *pb.Provider) (*pb.Provider, error) {
	//TODO implement me
	panic("implement me")
}
