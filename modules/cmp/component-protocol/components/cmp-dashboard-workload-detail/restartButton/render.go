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

package restartButton

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/cache"
	cputil2 "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	cmpTypes "github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/cmp/steve/middleware"
	"github.com/erda-project/erda/modules/cmp/steve/proxy"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "restartButton", func() servicehub.Provider {
		return &ComponentRestartButton{}
	})
}

var steveServer cmp.SteveServer

func (b *ComponentRestartButton) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return b.DefaultProvider.Init(ctx)
}

func (b *ComponentRestartButton) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	b.InitComponent(ctx)
	if err := b.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen operationButton component state, %v", err)
	}
	b.SetComponentValue()
	if event.Operation.String() == "restart" {
		if err := b.RestartWorkload(); err != nil {
			return errors.Errorf("failed to restart workload, %v", err)
		}
	}
	b.Transfer(component)
	return nil
}

func (b *ComponentRestartButton) InitComponent(ctx context.Context) {
	bdl := ctx.Value(cmpTypes.GlobalCtxKeyBundle).(*bundle.Bundle)
	b.bdl = bdl
	b.ctx = ctx
	sdk := cputil.SDK(ctx)
	b.sdk = sdk
	b.server = steveServer
}

func (b *ComponentRestartButton) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(component.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	b.State = state
	return nil
}

func (b *ComponentRestartButton) SetComponentValue() {
	splits := strings.Split(b.State.WorkloadID, "_")
	kind := splits[0]

	b.Props.Text = b.sdk.I18n("restart")
	b.Props.Type = "primary"

	operation := Operation{
		Key:     "restart",
		Reload:  true,
		Confirm: b.sdk.I18n("confirmRestart"),
	}
	if kind == string(apistructs.K8SJob) || kind == string(apistructs.K8SCronJob) {
		operation.Disabled = true
		operation.DisabledTip = b.sdk.I18n("cannotRestart")
	}

	b.Operations = map[string]interface{}{
		"click": operation,
	}
}

func (b *ComponentRestartButton) RestartWorkload() error {
	splits := strings.Split(b.State.WorkloadID, "_")
	if len(splits) != 3 {
		return errors.Errorf("invalid workload id, %s", b.State.WorkloadID)
	}

	kind, namespace, name := splits[0], splits[1], splits[2]
	if kind != string(apistructs.K8SDeployment) && kind != string(apistructs.K8SStatefulSet) &&
		kind != string(apistructs.K8SDaemonSet) {
		return errors.Errorf("invalid workload kind %s (only deployment, statefulSet and daemonSet can be restarted)", kind)
	}

	userID := b.sdk.Identity.UserID
	orgID := b.sdk.Identity.OrgID
	clusterName := b.State.ClusterName
	scopeID, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return apierrors.ErrInvoke.InvalidParameter(fmt.Errorf("invalid org id %s, %v", orgID, err))
	}

	if err := b.restartWorkload(userID, orgID, clusterName, kind, namespace, name); err != nil {
		return err
	}

	auditCtx := map[string]interface{}{
		middleware.AuditClusterName:  clusterName,
		middleware.AuditResourceName: name,
		middleware.AuditNamespace:    namespace,
		middleware.AuditResourceType: kind,
	}

	now := strconv.FormatInt(time.Now().Unix(), 10)
	if err := b.bdl.CreateAuditEvent(&apistructs.AuditCreateRequest{
		Audit: apistructs.Audit{
			UserID:       userID,
			ScopeType:    apistructs.OrgScope,
			ScopeID:      scopeID,
			OrgID:        scopeID,
			Context:      auditCtx,
			TemplateName: middleware.AuditRestartWorkload,
			Result:       "success",
			StartTime:    now,
			EndTime:      now,
		},
	}); err != nil {
		logrus.Errorf("failed to audit for restarting workload, %v", err)
	}
	return nil
}

func (b *ComponentRestartButton) restartWorkload(userID, orgID, clusterName, kind, namespace, name string) error {
	client, err := cputil2.GetImpersonateClient(b.server, userID, orgID, clusterName)
	if err != nil {
		return errors.Errorf("failed to get k8s client, %v", err)
	}

	patchBody := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/restartedAt": time.Now().Format("2006-01-02T15:04:05+07:00"),
					},
				},
			},
		},
	}

	data, err := json.Marshal(patchBody)
	if err != nil {
		return errors.Errorf("failed to marshal body, %v", err)
	}

	gvk := schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
	}
	switch kind {
	case string(apistructs.K8SDeployment):
		gvk.Kind = "Deployment"
		_, err = client.ClientSet.AppsV1().Deployments(namespace).Patch(b.ctx, name, types.StrategicMergePatchType, data, v1.PatchOptions{})
	case string(apistructs.K8SStatefulSet):
		gvk.Kind = "StatefulSet"
		_, err = client.ClientSet.AppsV1().StatefulSets(namespace).Patch(b.ctx, name, types.StrategicMergePatchType, data, v1.PatchOptions{})
	case string(apistructs.K8SDaemonSet):
		gvk.Kind = "DaemonSet"
		_, err = client.ClientSet.AppsV1().StatefulSets(namespace).Patch(b.ctx, name, types.StrategicMergePatchType, data, v1.PatchOptions{})
	default:
		return errors.Errorf("invalid workload kind %s (only deployment, statefulSet and daemonSet can be restarted)", kind)
	}
	if err != nil {
		return err
	}

	cacheKey := proxy.CacheKey{
		GVK:         gvk.String(),
		ClusterName: clusterName,
	}
	if _, err := cache.GetFreeCache().Remove(cacheKey.GetKey()); err != nil {
		logrus.Errorf("failed to remove cache for %s, %v", gvk.String(), err)
	}
	return nil
}

func (b *ComponentRestartButton) Transfer(component *cptype.Component) {
	component.Props = b.Props
	component.State = map[string]interface{}{
		"clusterName": b.State.ClusterName,
		"workloadId":  b.State.WorkloadID,
	}
	component.Operations = b.Operations
}
