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

package host

import (
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/container"
)

// Host 资源对象操作封装
type Host struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
	cs  *container.Container
}

// Option 定义 Host 对象的配置选项
type Option func(*Host)

// New 新建 Host 实例，通过 Host 实例操作主机资源
func New(options ...Option) *Host {
	h := &Host{}
	for _, op := range options {
		op(h)
	}
	return h
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(h *Host) {
		h.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(h *Host) {
		h.bdl = bdl
	}
}

// WithContainer 配置 container service
func WithContainer(cs *container.Container) Option {
	return func(h *Host) {
		h.cs = cs
	}
}

// CreateOrUpdate 创建或更新主机信息
func (h *Host) CreateOrUpdate(host *model.Host) error {
	old, err := h.db.GetHostByClusterAndIP(host.Cluster, host.PrivateAddr)
	if err != nil {
		return err
	}
	if old == nil {
		return h.db.CreateHost(host)
	}
	// 保持不变的信息
	host.ID = old.ID
	host.CreatedAt = old.CreatedAt

	return h.db.UpdateHost(host)
}

// Update 更新主机信息
func (h *Host) Update(host *model.Host) error {
	return h.db.UpdateHost(host)
}

// GetByClusterAndIP 根据 cluster & privateAddr获取主机信息
func (h *Host) GetByClusterAndIP(clusterName, privateAddr string) (*apistructs.Host, error) {
	host, err := h.db.GetHostByClusterAndIP(clusterName, privateAddr)
	if err != nil {
		return nil, err
	}
	return h.convert(host), nil
}

// GetByClusterAndPrivateIP 根据 cluster & privateAddr获取主机信息
func (h *Host) GetByClusterAndPrivateIP(clusterName, privateAddr string) (*model.Host, error) {
	return h.db.GetHostByClusterAndIP(clusterName, privateAddr)
}

// GetHostNumber 获取host数量
func (h *Host) GetHostNumber() (uint64, error) {
	return h.db.GetHostsNumber()
}

func (h *Host) convert(host *model.Host) *apistructs.Host {
	if host == nil {
		return nil
	}
	return &apistructs.Host{
		Name:          host.Name,
		OrgName:       host.OrgName,
		PrivateAddr:   host.PrivateAddr,
		Cpus:          host.Cpus,
		CpuUsage:      host.CpuUsage,
		Memory:        host.Memory,
		MemoryUsage:   host.MemoryUsage,
		Disk:          host.Disk,
		DiskUsage:     host.DiskUsage,
		Load5:         host.Load5,
		Cluster:       host.Cluster,
		Labels:        h.convertLegacyLabel(host.Labels),
		OS:            host.OS,
		KernelVersion: host.KernelVersion,
		SystemTime:    host.SystemTime,
		Birthday:      host.Birthday,
		TimeStamp:     host.TimeStamp,
		Deleted:       host.Deleted,
	}
}

// convertLegacyLabel 兼容marathon老数据，使新老标签展示一致
func (h *Host) convertLegacyLabel(labels string) string {
	labelSlice := strings.Split(labels, ",")
	newLabels := make([]string, 0, len(labelSlice))
	for _, v := range labelSlice {
		switch v {
		case "pack":
			newLabels = append(newLabels, "pack-job")
		case "bigdata":
			newLabels = append(newLabels, "bigdata-job")
		case "stateful", "service-stateful":
			newLabels = append(newLabels, "stateful-service")
		case "stateless", "service-stateless":
			newLabels = append(newLabels, "stateless-service")
		default:
			newLabels = append(newLabels, v)
		}
	}

	return strings.Join(newLabels, ",")
}
