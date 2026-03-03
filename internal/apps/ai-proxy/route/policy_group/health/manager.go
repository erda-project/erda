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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/lb/state_store"
	policygroup "github.com/erda-project/erda/internal/apps/ai-proxy/route/policy_group"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

type Config struct {
	ProbeBaseURL string        `file:"probe_base_url" env:"MODEL_HEALTH_PROBE_BASE_URL" default:"http://127.0.0.1:8081"`
	UnhealthyTTL time.Duration `file:"unhealthy_ttl" env:"MODEL_HEALTH_UNHEALTHY_TTL" default:"1h"`
	ProbeTimeout time.Duration `file:"probe_timeout" env:"MODEL_HEALTH_PROBE_TIMEOUT" default:"10s"`
	Rescue       RescueConfig  `file:"rescue"`
}

type RescueConfig struct {
	InitialBackoff time.Duration `file:"initial_backoff" env:"MODEL_HEALTH_RESCUE_INITIAL_BACKOFF" default:"3s"`
	MaxBackoff     time.Duration `file:"max_backoff" env:"MODEL_HEALTH_RESCUE_MAX_BACKOFF" default:"2m"`
}

var errUnsupportedAPIType = errors.New("health probe build request failed: unsupported api_type")

func (cfg *Config) normalize() {
	if cfg.ProbeBaseURL == "" {
		cfg.ProbeBaseURL = "http://127.0.0.1:8081"
	}
	if cfg.UnhealthyTTL <= 0 {
		cfg.UnhealthyTTL = time.Hour
	}
	if cfg.ProbeTimeout <= 0 {
		cfg.ProbeTimeout = 10 * time.Second
	}
	if cfg.Rescue.InitialBackoff <= 0 {
		panic("model health rescue initial_backoff must be > 0")
	}
	if cfg.Rescue.MaxBackoff <= 0 {
		panic("model health rescue max_backoff must be > 0")
	}
}

type workerState struct {
	mu      sync.Mutex
	apiType APIType
	headers http.Header
}

func (w *workerState) Update(apiType APIType, headers http.Header) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if apiType != "" {
		w.apiType = apiType
	}
	if len(headers) > 0 {
		w.headers = cloneHeaders(headers)
	}
}

func (w *workerState) Snapshot() (APIType, http.Header) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.apiType, cloneHeaders(w.headers)
}

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
		client: &http.Client{Timeout: cfg.ProbeTimeout},
	}
}

func (m *Manager) FilterHealthyInstances(req policygroup.RouteRequest, instances []*policygroup.RoutingModelInstance) []*policygroup.RoutingModelInstance {
	if m == nil || len(instances) == 0 {
		return instances
	}
	if isHealthProbeRequestMeta(req.Meta) {
		return instances
	}

	probeHeadersFromMeta := buildProbeHeadersFromRequestMeta(req.Meta)
	filtered := make([]*policygroup.RoutingModelInstance, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ModelWithProvider == nil {
			continue
		}
		state, ok, err := m.GetState(context.Background(), instance.ModelWithProvider.Id)
		if err != nil {
			return instances
		}
		if ok && strings.EqualFold(state.State, stateUnhealthy) {
			if !isAPITypeProbeSupported(state.APIType) {
				_ = m.store.DeleteBinding(context.Background(), modelHealthBindingKey, instance.ModelWithProvider.Id)
				if req.Ctx != nil {
					ctxhelper.AppendReleasedUnsupportedAPIType(req.Ctx, string(state.APIType))
				}
				filtered = append(filtered, instance)
				continue
			}
			if req.Ctx != nil {
				ctxhelper.AppendFilteredUnhealthyInstanceID(req.Ctx, instance.ModelWithProvider.Id)
			}
			// When process restarts, in-memory workers are gone but unhealthy state may still exist.
			// Re-arm probe worker from live user request context to avoid waiting for TTL.
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

func (m *Manager) startOrUpdateProbeWorker(instanceID string, apiType APIType, headers http.Header) {
	newState := &workerState{
		apiType: apiType,
		headers: cloneHeaders(headers),
	}
	actual, loaded := m.workers.LoadOrStore(instanceID, newState)
	if loaded {
		existing, _ := actual.(*workerState)
		if existing != nil {
			existing.Update(apiType, headers)
		}
		return
	}
	go m.probeWorker(instanceID, newState)
}

func (m *Manager) probeWorker(instanceID string, worker *workerState) {
	defer m.workers.Delete(instanceID)

	backoff := m.cfg.Rescue.InitialBackoff
	for {
		apiType, headers := worker.Snapshot()
		err := m.probeOnce(instanceID, apiType, headers)
		if err == nil {
			_ = m.store.DeleteBinding(context.Background(), modelHealthBindingKey, instanceID)
			return
		}
		if errors.Is(err, errUnsupportedAPIType) {
			// Unsupported api_type cannot be actively probed; fail-open to avoid
			// keeping the instance in unhealthy set indefinitely.
			_ = m.store.DeleteBinding(context.Background(), modelHealthBindingKey, instanceID)
			return
		}
		// Keep refreshing unhealthy state on every failed probe so the instance
		// is only released after a successful recovery probe.
		m.writeUnhealthyState(context.Background(), instanceID, apiType, err.Error())
		delay := withJitter(backoff)
		time.Sleep(delay)

		backoff = backoff * 2
		if backoff > m.cfg.Rescue.MaxBackoff {
			backoff = m.cfg.Rescue.MaxBackoff
		}
	}
}

func (m *Manager) probeOnce(instanceID string, apiType APIType, headers http.Header) error {
	path, body, ok := buildProbeRequest(apiType)
	if !ok {
		return errUnsupportedAPIType
	}

	baseURL := strings.TrimRight(m.cfg.ProbeBaseURL, "/")
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("health probe build http request failed: %w", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(vars.XAIProxyModelId, instanceID)
	req.Header.Set(vars.XAIProxyHealthProbe, "true")
	req.Header.Del("Content-Length")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("health probe request failed: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	// For network-failure recovery, receiving any HTTP response means the path is reachable.
	return nil
}

func (m *Manager) writeUnhealthyState(ctx context.Context, instanceID string, apiType APIType, lastErr string) {
	now := time.Now()
	state := ModelHealthState{
		State:     stateUnhealthy,
		APIType:   apiType,
		LastError: lastErr,
		UpdatedAt: now,
	}
	b, err := json.Marshal(&state)
	if err != nil {
		return
	}
	_ = m.store.SetBinding(ctx, modelHealthBindingKey, instanceID, string(b), m.cfg.UnhealthyTTL)
}

func buildProbeRequest(apiType APIType) (path string, body []byte, ok bool) {
	switch apiType {
	case APITypeChatCompletions:
		return vars.RequestPathPrefixV1ChatCompletions, []byte(`{"model":"health-probe","messages":[{"role":"user","content":"hello"}],"stream":false}`), true
	case APITypeResponses:
		return vars.RequestPathPrefixV1Responses, []byte(`{"model":"health-probe","input":"hello","stream":false}`), true
	default:
		return "", nil, false
	}
}

func isAPITypeProbeSupported(apiType APIType) bool {
	_, _, ok := buildProbeRequest(apiType)
	return ok
}

func isHealthProbeRequestMeta(meta policygroup.RequestMeta) bool {
	inputKeys := []string{
		common_types.StickyKeyPrefixFromReqHeader + strings.ToLower(vars.XAIProxyHealthProbe),
		strings.ToLower(vars.XAIProxyHealthProbe),
		vars.XAIProxyHealthProbe,
	}
	for _, key := range inputKeys {
		v, ok := meta.Get(key)
		if ok && strings.EqualFold(v, "true") {
			return true
		}
	}
	return false
}

func buildProbeHeadersFromRequestMeta(meta policygroup.RequestMeta) http.Header {
	headers := make(http.Header)
	for metaKey, metaValue := range meta.Keys {
		if !strings.HasPrefix(strings.ToLower(metaKey), common_types.StickyKeyPrefixFromReqHeader) {
			continue
		}
		headerKey := strings.TrimSpace(metaKey[len(common_types.StickyKeyPrefixFromReqHeader):])
		if headerKey == "" {
			continue
		}
		if strings.TrimSpace(metaValue) == "" {
			continue
		}
		headers.Set(http.CanonicalHeaderKey(headerKey), metaValue)
	}
	return headers
}

func cloneHeaders(headers http.Header) http.Header {
	if len(headers) == 0 {
		return http.Header{}
	}
	cloned := make(http.Header, len(headers))
	for key, values := range headers {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func withJitter(backoff time.Duration) time.Duration {
	if backoff <= 0 {
		return 0
	}
	jitterRange := backoff / 4
	if jitterRange <= 0 {
		return backoff
	}
	delta := time.Duration(rand.Int63n(int64(jitterRange)))
	return backoff + delta
}

func (m *Manager) String() string {
	if m == nil {
		return "nil-health-manager"
	}
	return fmt.Sprintf("health-manager(base=%s)", m.cfg.ProbeBaseURL)
}
