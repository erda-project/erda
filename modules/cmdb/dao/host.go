package dao

import (
	"context"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateHost 创建host
func (client *DBClient) CreateHost(host *model.Host) error {
	return client.Create(host).Error
}

//UpdateHost 更新host
func (client *DBClient) UpdateHost(host *model.Host) error {
	return client.Save(host).Error
}

// GetHostByClusterAndIP 根据 cluster & privateAddr获取主机信息
func (client *DBClient) GetHostByClusterAndIP(clusterName, privateAddr string) (*model.Host, error) {
	var host model.Host
	if err := client.Where("cluster = ?", clusterName).
		Where("private_addr = ?", privateAddr).First(&host).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return &host, nil
}

// GetHostsByCluster 根据 clusterName获取主机列表
func (client *DBClient) GetHostsByCluster(clusterName string) (*[]model.Host, error) {
	var hosts []model.Host
	if err := client.Where("cluster = ?", clusterName).Find(&hosts).Error; err != nil {
		return nil, err
	}
	return &hosts, nil
}

// GetHostsByClusterAndNullOrg 为兼容dcos 有些机器未加 org 标
func (client *DBClient) GetHostsByClusterAndNullOrg(clusterName string) (*[]model.Host, error) {
	var hosts []model.Host
	if err := client.Where("cluster = ?", clusterName).Where("org_name = ''").
		Find(&hosts).Error; err != nil {
		return nil, err
	}
	return &hosts, nil
}

// GetHostsByClusterAndOrg 根据 clusterName & orgName 获取主机列表
func (client *DBClient) GetHostsByClusterAndOrg(clusterName, orgName string) (*[]model.Host, error) {
	var hosts []model.Host
	if err := client.Where("cluster = ?", clusterName).
		Where("(org_name = ? or org_name LIKE ? or org_name LIKE ? or org_name LIKE ?)",
			orgName, strutil.Concat(orgName, ",%"), strutil.Concat("%,", orgName), strutil.Concat("%,", orgName, ",%")).
		Find(&hosts).Error; err != nil {
		return nil, err
	}
	return &hosts, nil
}

// QueryHost 获取单个host的信息
func (client *DBClient) QueryHost(ctx context.Context, cluster, addr string) (*types.CmHost, error) {
	var h types.CmHost
	var err error

	if cluster == "" || addr == "" {
		return nil, errors.Errorf("invalid params: cluster = %s, addr = %s", cluster, addr)
	}

	if err = client.Where("cluster = ? AND private_addr = ?", cluster, addr).Find(&h).Error; err != nil {
		return nil, err
	}

	return &h, nil
}

// AllHostsByCluster 获取指定集群下所有的host信息
func (client *DBClient) AllHostsByCluster(ctx context.Context, cluster string) (*[]types.CmHost, error) {
	var hosts []types.CmHost
	var err error

	if cluster == "" {
		return nil, errors.Errorf("invalid params: cluster = %s", cluster)
	}

	if err = client.Where("cluster = ?", cluster).Find(&hosts).Error; err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	return &hosts, nil
}

// DeleteHostByName 根据名字删除指定集群下指定的宿主机
func (client *DBClient) DeleteHostByName(ctx context.Context, cluster, name string) error {

	if cluster == "" || name == "" {
		return errors.Errorf("invalid params: cluster = %s, host.name = %s", cluster, name)
	}

	return client.Where("cluster = ? AND name = ?", cluster, name).Delete(types.CmHost{}).Error
}

// TODO deprecated
// MakeHostLabel 处理宿主机调度标签可读
func MakeHostLabel(hostLabels string) string {
	var diceLabels string

	if len(hostLabels) == 0 {
		return ""
	}

	allLabels := strings.Split(hostLabels, ";")
	for _, label := range allLabels {
		if strings.Contains(label, "dice_tags:") {
			diceLabels = strings.Split(label, "\n")[0]
			break
		}
	}

	if len(diceLabels) == 0 {
		diceLabels = "dice_tags:"
	}

	currentLabel := strings.Split(diceLabels, ":")[1]
	labels := strings.Split(currentLabel, ",")

	var newLabels, anyLabel, lockLabel, otherLabels []string
	for _, label := range labels {
		label = convertLegacyLabel(label)
		if label == "any" {
			anyLabel = append(anyLabel, label)
		} else if label == "locked" {
			lockLabel = append(lockLabel, label)
		} else {
			otherLabels = append(otherLabels, label)
		}
	}

	newLabels = append(newLabels, lockLabel...)
	newLabels = append(newLabels, anyLabel...)
	newLabels = append(newLabels, otherLabels...)

	return strings.Join(newLabels, ",")
}

// TODO deprecated
func convertLegacyLabel(label string) string {
	switch label {
	case "pack":
		return "pack-job"
	case "bigdata":
		return "bigdata-job"
	case "stateful", "service-stateful":
		return "stateful-service"
	case "stateless", "service-stateless":
		return "stateless-service"
	default:
		return label
	}
}

// GetHostsNumberByClusterAndOrg 根据 clusterName & orgName 获取集群主机总数
func (client *DBClient) GetHostsNumberByClusterAndOrg(clusterName, orgName string) (uint64, error) {
	var count uint64
	if err := client.Model(&model.Host{}).
		Where("cluster = ?", clusterName).
		Where("org_name = ?", orgName).
		Select("count(distinct(private_addr))").Count(&count).
		Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetHostsNumber 查询host数量
func (client *DBClient) GetHostsNumber() (uint64, error) {
	var count uint64
	if err := client.Model(&model.Host{}).
		Select("count(distinct(private_addr))").Count(&count).
		Error; err != nil {
		return 0, err
	}
	return count, nil
}

// GetAbnormalHostsNumberByClusterAndOrg 根据 clusterName & orgName 获取集群异常主机总数
func (client *DBClient) GetAbnormalHostsNumberByClusterAndOrg(clusterName, orgName string) (uint64, error) {
	var count uint64
	if err := client.Model(&model.Host{}).
		Where("cluster = ?", clusterName).
		Where("(org_name = ? or org_name LIKE ? or org_name LIKE ? or org_name LIKE ?)",
			orgName, strutil.Concat(orgName, ",%"), strutil.Concat("%,", orgName), strutil.Concat("%,", orgName, ",%")).
		Select("count(distinct(private_addr))").Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
