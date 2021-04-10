package assetsvc

import (
	"github.com/erda-project/erda/apistructs"
)

func onceADayLimitType() []apistructs.LimitType {
	once := 1
	return []apistructs.LimitType{{
		Day:    &once,
		Hour:   nil,
		Minute: nil,
		Second: nil,
	}}
}

func unlimitedSLA(access *apistructs.APIAccessesModel) *apistructs.SLAModel {
	return &apistructs.SLAModel{
		BaseModel: apistructs.BaseModel{
			ID:        0,
			CreatedAt: access.CreatedAt,
			UpdatedAt: access.CreatedAt,
			CreatorID: access.CreatorID,
			UpdaterID: access.CreatorID,
		},
		Name:     apistructs.UnlimitedSLAName,
		Desc:     "不限制流量 SLA",
		Approval: apistructs.AuthorizationManual,
		AccessID: access.ID,
		Source:   apistructs.SourceSystem,
	}
}

// 如果传入的 slaID 为 0, 返回无限制流量的 SLA; 否则查询数据库返回 slaID 对应的库表记录
func (svc *Service) querySLAByID(slaID uint64, access *apistructs.APIAccessesModel) (*apistructs.SLAModel, error) {
	if slaID == 0 {
		return unlimitedSLA(access), nil
	}
	var sla apistructs.SLAModel
	if err := svc.FirstRecord(&sla, map[string]interface{}{"id": slaID}); err != nil {
		return nil, err
	}
	return &sla, nil
}
