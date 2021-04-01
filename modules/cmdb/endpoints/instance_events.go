// Package endpoints cmdb api 逻辑处理层
package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/cmdb/types"
	"github.com/erda-project/erda/pkg/httpserver"
)

var runtimeNamePrefix = "services/"

func (e *Endpoints) UpdateInstanceBySchedulerEvent(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var event apistructs.InstanceStatusEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		return apierrors.ErrSyncInstance.InvalidParameter(err).ToResp(), nil
	}
	logrus.Debugf("get instance event from scheduler: %+v", event)

	if err := checkInstanceEvent(event.Content); err != nil {
		return apierrors.ErrSyncInstance.InvalidParameter(err).ToResp(), nil
	}

	c, err := fillCmContainerByEvent(&event)
	if err != nil {
		return apierrors.ErrSyncInstance.InternalError(err).ToResp(), nil
	}

	etcdKey := fmt.Sprintf("/schedulerevent/%s-%s", c.TaskID, strings.ToLower(c.Status))
	// 确保多个实例任意时刻只有一个执行
	txResp, err := e.etcdStore.GetClient().Txn(context.Background()).
		If(v3.Compare(v3.Version(etcdKey), "=", 0)).
		Then(v3.OpPut(etcdKey, "true")).
		Commit()
	if err != nil || !txResp.Succeeded {
		logrus.Warnf("failed to get lock: %s", etcdKey)
		return httpserver.OkResp(nil)
	}

	// 根据 taskID 进行查询
	containers, err := e.container.GetContainerByTaskIDOrContainerID(event.Content.ClusterName, event.Content.ID, "")
	if err != nil {
		return apierrors.ErrSyncInstance.InternalError(err).ToResp(), nil
	}
	// 若未找到，则新增
	if len(containers) == 0 {
		if err := e.container.Create(c); err != nil {
			return apierrors.ErrSyncInstance.InternalError(err).ToResp(), nil
		}
	} else {
		oldContainer := containers[0]
		oldContainer.TaskID = c.TaskID
		if c.IPAddress != "" {
			oldContainer.IPAddress = c.IPAddress
		}
		if c.HostPrivateIPAddr != "" {
			oldContainer.HostPrivateIPAddr = c.HostPrivateIPAddr
		}
		oldContainer.Status = c.Status
		oldContainer.DiceRuntime = c.DiceRuntime
		oldContainer.DiceService = c.DiceService
		if err := e.container.Update(&oldContainer); err != nil {
			return apierrors.ErrSyncInstance.InternalError(err).ToResp(), nil
		}
	}

	if _, err := e.etcdStore.GetClient().Delete(context.Background(), etcdKey); err != nil {
		logrus.Warnf("failed to release lock: %s", etcdKey)
	}

	return httpserver.OkResp(nil)
}

func checkInstanceEvent(event apistructs.InstanceStatusData) error {
	if event.ID == "" {
		return errors.Errorf("invalid scheduler event: task id is empty.")
	}

	if event.ClusterName == "" {
		return errors.Errorf("invalid scheduler event: cluster name is empty.")
	}

	if !strings.HasPrefix(event.RuntimeName, runtimeNamePrefix) {
		return errors.Errorf("invalid runtime name(%s) from scheduler event, no %s prefix.", event.RuntimeName, runtimeNamePrefix)
	}

	if event.ServiceName == "" {
		return errors.Errorf("invalid scheduler event: service name is empty.")
	}

	if !types.IsValidSchedulerInstanceStatus(event.InstanceStatus) {
		return errors.Errorf("ignore instance status from scheduler event: %s", event.InstanceStatus)
	}

	return nil
}

func fillCmContainerByEvent(event *apistructs.InstanceStatusEvent) (*model.Container, error) {
	var (
		workspace string
		runtimeID string
		err       error
	)

	content := event.Content
	workspace, runtimeID, err = splitRuntimeNameToIDAndWorkspace(content.RuntimeName)
	if err != nil {
		return nil, err
	}

	c := &model.Container{
		TaskID:            content.ID,
		Cluster:           content.ClusterName,
		IPAddress:         content.IP,
		HostPrivateIPAddr: content.Host,
		DiceWorkspace:     workspace,
		DiceRuntime:       runtimeID,
		DiceService:       content.ServiceName,
		Status:            content.InstanceStatus,
		TimeStamp:         content.Timestamp,
	}

	switch types.InstanceStatus(content.InstanceStatus) {
	case types.InstanceStatusKilled, types.InstanceStatusFailed, types.InstanceStatusFinished:
		c.FinishedAt = time.Now().Format(time.RFC3339Nano)
	case types.InstanceStatusHealthy, types.InstanceStatusUnHealthy:
		if startAt, err := time.Parse(time.RFC3339Nano, event.TimeStamp); err == nil {
			c.StartedAt = startAt.String()
		}
	}

	return c, nil
}

func splitRuntimeNameToIDAndWorkspace(runtimeName string) (string, string, error) {
	var tmpName string
	var workspace, runtimeID string

	// e.g. runtimeName: services/dev-111
	tmpName = strings.TrimPrefix(runtimeName, runtimeNamePrefix)
	body := strings.Split(tmpName, "-")
	if len(body) != 2 {
		return "", "", errors.Errorf("invalid runtime name: %s", runtimeName)
	}

	workspace = body[0]
	runtimeID = body[1]

	return workspace, runtimeID, nil
}
