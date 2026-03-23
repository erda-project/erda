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
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

var errUnsupportedAPIType = errors.New("health probe build request failed: unsupported api_type")

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

func (m *Manager) startOrUpdateProbeWorker(clientID, instanceID string, apiType APIType, headers http.Header) {
	newState := &workerState{
		apiType: apiType,
		headers: cloneHeaders(headers),
	}
	requestID := extractRequestID(headers)
	workerKey := makeModelHealthBindingID(clientID, instanceID)
	actual, loaded := m.workers.LoadOrStore(workerKey, newState)
	if loaded {
		existing, _ := actual.(*workerState)
		if existing != nil {
			existing.Update(apiType, headers)
		}
		logrus.WithFields(logrus.Fields{
			"client_id":   clientID,
			"instance_id": instanceID,
			"api_type":    apiType,
			"request_id":  requestID,
		}).Warn("probe worker already running, update worker state")
		return
	}
	logrus.WithFields(logrus.Fields{
		"client_id":   clientID,
		"instance_id": instanceID,
		"api_type":    apiType,
		"request_id":  requestID,
	}).Warn("start probe worker for unhealthy instance")
	go m.probeWorker(clientID, instanceID, newState)
}

func (m *Manager) probeWorker(clientID, instanceID string, worker *workerState) {
	workerKey := makeModelHealthBindingID(clientID, instanceID)
	defer m.workers.Delete(workerKey)

	backoff := m.cfg.Rescue.InitialBackoff
	for {
		apiType, headers := worker.Snapshot()
		requestID, err := m.probeOnce(instanceID, apiType, headers)
		if err == nil {
			_ = m.store.DeleteBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID(clientID, instanceID))
			logrus.WithFields(logrus.Fields{
				"client_id":   clientID,
				"instance_id": instanceID,
				"api_type":    apiType,
				"request_id":  requestID,
			}).Warn("probe success, recovered unhealthy instance")
			return
		}
		if errors.Is(err, errUnsupportedAPIType) {
			_ = m.store.DeleteBinding(context.Background(), modelHealthBindingKey, makeModelHealthBindingID(clientID, instanceID))
			logrus.WithFields(logrus.Fields{
				"client_id":   clientID,
				"instance_id": instanceID,
				"api_type":    apiType,
				"request_id":  requestID,
			}).Warnf("unsupported api_type for probe, release unhealthy instance immediately, err: %v", err)
			return
		}
		m.writeUnhealthyState(context.Background(), clientID, instanceID, apiType, err.Error())
		logrus.WithFields(logrus.Fields{
			"client_id":   clientID,
			"instance_id": instanceID,
			"api_type":    apiType,
			"backoff":     backoff.String(),
			"request_id":  requestID,
		}).Warnf("probe failed, keep unhealthy and retry, err: %v", err)
		delay := withJitter(backoff)
		time.Sleep(delay)

		backoff = backoff * 2
		if backoff > m.cfg.Rescue.MaxBackoff {
			backoff = m.cfg.Rescue.MaxBackoff
		}
	}
}

func (m *Manager) probeOnce(instanceID string, apiType APIType, headers http.Header) (string, error) {
	path, body, ok := buildProbeRequest(apiType)
	if !ok {
		return extractRequestID(headers), errUnsupportedAPIType
	}

	baseURL := strings.TrimRight(m.cfg.Probe.BaseURL, "/")
	req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return extractRequestID(headers), fmt.Errorf("health probe build http request failed: %w", err)
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(vars.XAIProxyModelId, instanceID)
	req.Header.Set(vars.XAIProxyModelHealthProbe, "true")
	req.Header.Del("Content-Length")

	resp, err := m.client.Do(req)
	if err != nil {
		return extractRequestID(req.Header), fmt.Errorf("health probe request failed: %w", err)
	}
	defer resp.Body.Close()
	requestID := extractRequestID(resp.Header)
	if requestID == "" {
		requestID = extractRequestID(req.Header)
	}
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return requestID, fmt.Errorf("health probe got non-2xx status=%d body=%q", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return requestID, nil
}

func (m *Manager) writeUnhealthyState(ctx context.Context, clientID, instanceID string, apiType APIType, lastErr string) {
	m.writeUnhealthyStateAt(ctx, clientID, instanceID, apiType, lastErr, time.Now())
}

func (m *Manager) writeUnhealthyStateAt(ctx context.Context, clientID, instanceID string, apiType APIType, lastErr string, now time.Time) {
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
	_ = m.store.SetBinding(ctx, modelHealthBindingKey, makeModelHealthBindingID(clientID, instanceID), string(b), m.cfg.Probe.UnhealthyTTL)
}
