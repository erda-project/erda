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

package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type Manager struct {
	store   state_store.LBStateStore
	client  *http.Client
	cfg     Config
	workers sync.Map // map[instanceID]*workerState
}

func NewManager(store state_store.LBStateStore, cfg Config) *Manager {
	if store == nil {
		return nil
	}
	cfg.normalize()
	return &Manager{
		store:  store,
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Probe.Timeout},
	}
}

func (m *Manager) FilterHealthyInstances(req policygroup.RouteRequest, instances []*policygroup.RoutingModelInstance) []*policygroup.RoutingModelInstance {
	if m == nil || len(instances) == 0 {
		return instances
	}
	if req.Ctx != nil {
		if trusted, ok := ctxhelper.GetTrustedHealthProbe(req.Ctx); ok && trusted {
			return instances
		}
	}

	probeHeadersFromMeta := buildProbeHeadersFromRequestMeta(req.Meta)
	if req.Ctx != nil {
		if reqID, ok := ctxhelper.GetRequestID(req.Ctx); ok && reqID != "" {
			if probeHeadersFromMeta.Get(vars.XRequestId) == "" {
				probeHeadersFromMeta.Set(vars.XRequestId, reqID)
			}
		}
	}
	storeCtx := req.Ctx
	if storeCtx == nil {
		storeCtx = context.Background()
	}
	filtered := make([]*policygroup.RoutingModelInstance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := m.GetState(storeCtx, instance.ModelWithProvider.Id)
		if err != nil {
			return instances
		}
		if ok && strings.EqualFold(state.State, stateUnhealthy) {
			if !isAPITypeProbeSupported(state.APIType) {
				_ = m.store.DeleteBinding(storeCtx, modelHealthBindingKey, instance.ModelWithProvider.Id)
				if req.Ctx != nil {
					AppendReleasedUnsupportedAPIType(req.Ctx, string(state.APIType))
				}
				filtered = append(filtered, instance)
				continue
			}
			if req.Ctx != nil {
				AppendFilteredUnhealthyInstanceID(req.Ctx, instance.ModelWithProvider.Id)
			}
			m.startOrUpdateProbeWorker(instance.ModelWithProvider.Id, state.APIType, probeHeadersFromMeta)
			continue
		}
		filtered = append(filtered, instance)
	}
	return filtered
}

func (m *Manager) MarkUnhealthy(ctx context.Context, instanceID string, apiType APIType, lastErr string, headers http.Header) {
	if m == nil || instanceID == "" || apiType == "" {
		return
	}
	if headers == nil {
		headers = make(http.Header)
	}
	if ctx != nil && headers.Get(vars.XRequestId) == "" {
		if reqID, ok := ctxhelper.GetRequestID(ctx); ok && reqID != "" {
			headers.Set(vars.XRequestId, reqID)
		}
	}
	callID := extractCallID(headers)
	logrus.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"api_type":    apiType,
		"error":       lastErr,
		"call_id":     callID,
	}).Warn("mark model instance unhealthy")
	tryPutModelMarkUnhealthyInstanceID(ctx, instanceID)
	m.writeUnhealthyState(ctx, instanceID, apiType, lastErr)
	m.startOrUpdateProbeWorker(instanceID, apiType, headers)
}

func (m *Manager) GetState(ctx context.Context, instanceID string) (*ModelHealthState, bool, error) {
	val, ok, err := m.store.GetBinding(ctx, modelHealthBindingKey, instanceID)
	if err != nil || !ok || val == "" {
		return nil, ok, err
	}
	var state ModelHealthState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, false, err
	}
	return &state, true, nil
}

func (m *Manager) String() string {
	if m == nil {
		return "nil-health-manager"
	}
	return fmt.Sprintf("health-manager(base=%s)", m.cfg.Probe.BaseURL)
}
