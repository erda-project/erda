package block

import (
	"fmt"

	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/modules/pkg/mysql"
	"github.com/erda-project/erda/pkg/uuid"
)

type DashboardAPI interface {
	CreateDashboard(block *UserBlock) (dash *DashboardBlockDTO, err error)
	// GenerateDashboardWithViewAPIList(api []*API) (block *UserBlock, err error)
}

func (p *provider) CreateDashboard(body *UserBlock) (dash *DashboardBlockDTO, err error) {
	if len(body.ID) == 0 {
		body.ID = uuid.UUID()
	}
	if body.DataConfig == nil {
		body.DataConfig = &dataConfigDTO{}
	}
	if err := p.db.userBlock.Save(body); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return nil, fmt.Errorf("aleady existed, err: %s", err)
		}
		return nil, err
	}
	if body.ViewConfig != nil && body.DataConfig != nil {
		for _, v := range *body.ViewConfig {
			v.View.StaticData = struct{}{}
			for _, d := range *body.DataConfig {
				if v.I == d.I {
					v.View.StaticData = d.StaticData
				}
			}
		}
	}
	return &DashboardBlockDTO{
		ID:         body.ID,
		Name:       body.Name,
		Desc:       body.Desc,
		Scope:      body.Scope,
		ScopeID:    body.ScopeID,
		ViewConfig: body.ViewConfig,
		CreatedAt:  utils.ConvertTimeToMS(body.CreatedAt),
		UpdatedAt:  utils.ConvertTimeToMS(body.UpdatedAt),
		Version:    body.Version,
	}, nil
}
