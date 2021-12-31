package deployment_order

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
)

func (d *DeploymentOrder) List(projectId uint64, pageInfo *apistructs.PageInfo) (*apistructs.DeploymentOrderListData, error) {
	total, data, err := d.db.ListDeploymentOrder(projectId, pageInfo)
	if err != nil {
		return nil, err
	}

	return &apistructs.DeploymentOrderListData{
		Total: total,
		List:  convertDeploymentOrderToResponseItem(data),
	}, nil
}

func convertDeploymentOrderToResponseItem(orders []dbclient.DeploymentOrder) []*apistructs.DeploymentOrderItem {
	ret := make([]*apistructs.DeploymentOrderItem, 0)

	for _, order := range orders {
		ret = append(ret, &apistructs.DeploymentOrderItem{
			ID:        order.ID,
			Name:      order.Name,
			ReleaseID: order.ReleaseId,
			Params:    order.Params, // TODO: parse response format
			Type:      order.Type,
			Status:    apistructs.DeploymentOrderStatus(order.Status),
			Operator:  string(order.Operator),
		})
	}

	return ret
}
