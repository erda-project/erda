package dbclient

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/apistructs"
)

const (
	orderTableName = "erda_deployment_order"
)

type DeploymentOrder struct {
	ID              string `gorm:"size:36"`
	Name            string
	Source          string // TODO: deprecated
	Type            string
	Desc            string
	ReleaseId       string
	Operator        user.ID `gorm:"not null;"`
	ProjectId       uint64
	ProjectName     string
	ApplicationId   int64
	ApplicationName string
	Status          string
	Params          string
	Outdated        uint16
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (DeploymentOrder) TableName() string {
	return orderTableName
}

func (db *DBClient) ListDeploymentOrder(projectId uint64, pageInfo *apistructs.PageInfo) (int, []DeploymentOrder, error) {
	cursor := db.Where("project_id = ?", projectId)

	var (
		total  int
		orders = make([]DeploymentOrder, 0)
	)

	if err := cursor.Order("created_at desc").Offset(pageInfo.GetOffset()).Limit(pageInfo.GetLimit()).Find(&orders).
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return 0, nil, fmt.Errorf("failed to list deployment order, projectId: %d, err: %v", pageInfo, err)
	}

	return total, orders, nil
}

func (db *DBClient) GetOrderCountByProject(projectId uint64, tp string) (int64, error) {
	if tp == apistructs.TypePipeline {
		return 0, fmt.Errorf("pipeline type doesn't need to count")
	}

	var count int64

	if err := db.Model(&DeploymentOrder{}).Where("project_id = ? and type = ?", projectId, tp).Count(&count).Error; err != nil {
		return 0, errors.Wrapf(err, "failed to count, project: %d, rg: %s", projectId, tp)
	}

	return count, nil
}

func (db *DBClient) GetOrCreateDeploymentOrder(do *DeploymentOrder) error {
	// TODO
	//db.Find(&DeploymentOrderItem{
	//	Name:            do.Name,
	//	ProjectId:       do.ProjectId,
	//	ApplicationId:   do.ApplicationId,
	//	ApplicationName: "",
	//	Status:          "",
	//	Params:          "",
	//	Outdated:        0,
	//	CreatedAt:       time.Time{},
	//	UpdatedAt:       time.Time{},
	//})

	if err := db.Save(do).Error; err != nil {
		return errors.Wrapf(err, "failed to create deployment order, error: %s", err)
	}
	return nil
}
