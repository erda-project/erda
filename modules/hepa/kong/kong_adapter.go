// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package kong

import (
	"net/http"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/kong/base"
	v2 "github.com/erda-project/erda/modules/hepa/kong/v2"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	"github.com/erda-project/erda/modules/hepa/repository/service"
)

const InnerHost = "gateway.inner"

const (
	ConsumerRoot string = "/consumers/"
	PluginRoot   string = "/plugins/"
	ServiceRoot  string = "/services/"
	RouteRoot    string = "/routes/"
	AclRoot      string = "/acls/"
	UpstreamRoot string = "/upstreams/"
	HealthPath   string = "/health/"
	TargetPath   string = "/targets/"
)

var (
	ErrInvalidReq = errors.New("kongAdapter: invalid request")
)

func newKongAdapter(kongAddr string, client *http.Client) KongAdapter {
	var empty *base.KongAdapterImpl
	base := &base.KongAdapterImpl{
		KongAddr: kongAddr,
		Client:   client,
	}
	version, err := base.GetVersion()
	if err != nil {
		log.Errorf("get kong version failed, addr:%s", kongAddr)
		return empty
	}
	if strings.HasPrefix(version, "2.") {
		return &v2.KongAdapterImpl{
			KongAdapterImpl: base,
		}
	} else if strings.HasPrefix(version, "0.") {
		return base
	}
	log.Errorf("not support version:%s", version)
	return empty
}

func NewKongAdapter(kongAddr string) KongAdapter {
	client := &http.Client{}
	if config.ServerConf.KongDebug {
		return newKongAdapter(config.ServerConf.KongDebugAddr, client)
	}
	return newKongAdapter(kongAddr, client)
}

func NewKongAdapterForConsumer(consumer *orm.GatewayConsumer) KongAdapter {
	client := &http.Client{}
	if config.ServerConf.KongDebug {
		return newKongAdapter(config.ServerConf.KongDebugAddr, client)
	}
	azInfoService, err := service.NewGatewayAzInfoServiceImpl()
	if err != nil {
		log.Error(err)
		return nil
	}
	az, err := azInfoService.GetAz(&orm.GatewayAzInfo{
		OrgId:     consumer.OrgId,
		ProjectId: consumer.ProjectId,
		Env:       consumer.Env,
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	return NewKongAdapterForProject(az, consumer.Env, consumer.ProjectId)
}

func NewKongAdapterForProject(az, env, projectId string) KongAdapter {
	client := &http.Client{}
	if config.ServerConf.KongDebug {
		return newKongAdapter(config.ServerConf.KongDebugAddr, client)
	}
	kongInfoService, err := service.NewGatewayKongInfoServiceImpl()
	if err != nil {
		log.Error(err)
		return nil
	}
	kong, err := kongInfoService.GetKongInfo(&orm.GatewayKongInfo{
		Az:        az,
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		log.Error(err)
		return nil
	}
	if kong == nil {
		log.Error("can't find kong")
		return nil
	}
	return newKongAdapter(kong.KongAddr, client)
}

func NewKongAdapterByConsumerId(consumerDb service.GatewayConsumerService, consumerId string) KongAdapter {
	client := &http.Client{}
	if config.ServerConf.KongDebug {
		return newKongAdapter(config.ServerConf.KongDebugAddr, client)
	}
	consumer, err := consumerDb.GetById(consumerId)
	if err != nil {
		log.Errorf("consumerDb failed[%+v]", errors.WithStack(err))
		return nil
	}
	if consumer == nil {
		log.Errorf("consumer[%s] not exists", consumerId)
		return nil
	}
	return NewKongAdapterForConsumer(consumer)
}
