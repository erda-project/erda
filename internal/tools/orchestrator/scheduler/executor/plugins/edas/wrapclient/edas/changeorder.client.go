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

package edas

import (
	"time"

	api "github.com/aliyun/alibaba-cloud-sdk-go/services/edas"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/types"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/executor/plugins/edas/utils"
)

var loopCount = 2 * 60

// ListRecentChangeOrderInfo Query the list of release history
func (c *wrapEDAS) ListRecentChangeOrderInfo(appID string) (*api.ChangeOrderList, error) {
	l := c.l.WithField("func", "ListRecentChangeOrderInfo")

	req := api.CreateListRecentChangeOrderRequest()
	req.SetDomain(c.addr)
	req.AppId = appID

	resp, err := c.client.ListRecentChangeOrder(req)
	if err != nil {
		return nil, errors.Wrap(err, "edas list recent change order info")
	}

	l.Info("request id", resp.RequestId)

	return &resp.ChangeOrderList, nil
}

// LoopTerminationStatus Get the results of the task release list in a loop
func (c *wrapEDAS) LoopTerminationStatus(orderID string) (types.ChangeOrderStatus, error) {
	var status types.ChangeOrderStatus
	var err error

	l := c.l.WithField("func", "LoopTerminationStatus")
	l.Infof("start to loop termination status, order id: %v", orderID)

	retry := 2
	for i := 0; i < loopCount; i++ {
		if i > 0 {
			time.Sleep(10 * time.Second)
		}

		status, err = c.getChangeOrderInfo(orderID)
		if err != nil {
			return status, err
		}

		if status == types.CHANGE_ORDER_STATUS_PENDING || status == types.CHANGE_ORDER_STATUS_EXECUTING {
			continue
		}

		if status == types.CHANGE_ORDER_STATUS_SUCC || retry <= 0 {
			return status, nil
		}
		retry--
	}

	return status, errors.Errorf("get change order info timeout, order id: %s.", orderID)
}

// AbortChangeOrder Termination of change order
func (c *wrapEDAS) AbortChangeOrder(changeOrderID string) error {
	l := c.l.WithField("func", "AbortChangeOrder")

	l.Infof("start to abort change order, id: %s", changeOrderID)

	req := api.CreateAbortChangeOrderRequest()
	req.Headers = utils.AppendCommonHeaders(req.Headers)
	req.SetDomain(c.addr)
	req.ChangeOrderId = changeOrderID

	resp, err := c.client.AbortChangeOrder(req)
	if err != nil {
		return errors.Errorf("failed to abort change order(%s), err: %v", changeOrderID, err)
	}

	l.Info("request id", resp.RequestId)

	return nil
}

// getChangeOrderInfo Check details of changes
func (c *wrapEDAS) getChangeOrderInfo(orderID string) (types.ChangeOrderStatus, error) {
	l := c.l.WithField("func", "getChangeOrderInfo")
	l.Debugf("start to get change order info, orderID: %s", orderID)

	req := api.CreateGetChangeOrderInfoRequest()
	req.SetDomain(c.addr)
	req.ChangeOrderId = orderID

	resp, err := c.client.GetChangeOrderInfo(req)
	if err != nil {
		return types.CHANGE_ORDER_STATUS_ERROR, errors.Wrap(err, "edas get change order info")
	}

	l.Info("request id", resp.RequestId)

	status := types.ChangeOrderStatus(resp.ChangeOrderInfo.Status)
	l.Infof("get change order info, orderID: %s, status: %v", orderID, status)

	return status, nil
}
