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

package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/tools/gittar/pkg/admin_token"
	"github.com/erda-project/erda/internal/tools/gittar/rpcmetrics"
	"github.com/erda-project/erda/internal/tools/gittar/webcontext"
)

type RPCMetricsResponse struct {
	Enabled           bool               `json:"enabled"`
	FileWriterEnabled bool               `json:"fileWriterEnabled"`
	FilePath          string             `json:"filePath"`
	InMemory          RPCMetricsInMemory `json:"in_memory"`
	Summary           rpcmetrics.Summary `json:"summary"`
}

type RPCMetricsInMemory struct {
	Counters rpcmetrics.Counters       `json:"counters"`
	Active   rpcmetrics.ActiveSnapshot `json:"active"`
}

func RPCMetrics(ctx *webcontext.Context) {
	if !verifyAdminAuthToken(ctx) {
		return
	}

	limitActive := ctx.GetQueryInt32("limit_active", 100)
	if limitActive < 0 {
		limitActive = 0
	}
	if limitActive > 1000 {
		limitActive = 1000
	}

	limitTopStr := ctx.Query("limit_top")
	limitTop := 10
	if limitTopStr != "" {
		if val, err := strconv.Atoi(limitTopStr); err == nil {
			limitTop = val
		}
	}
	if limitTop < 0 {
		limitTop = 0
	}
	if limitTop > 1000 {
		limitTop = 1000
	}

	minDurationMS, _ := strconv.ParseInt(ctx.Query("min_duration_ms"), 10, 64)
	if minDurationMS < 0 {
		minDurationMS = 0
	}

	service := ctx.Query("service")
	phase := ctx.Query("phase")

	dateStr := ctx.Query("date")
	targetDate := time.Now()
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = t
		}
	}

	summary := rpcmetrics.BuildSummary(rpcmetrics.GetFilePath(targetDate), limitTop)
	resp := RPCMetricsResponse{
		Enabled:           rpcmetrics.Enabled(),
		FileWriterEnabled: rpcmetrics.FileWriterEnabled(),
		FilePath:          rpcmetrics.GetFilePath(targetDate),
		InMemory: RPCMetricsInMemory{
			Counters: rpcmetrics.GetCounters(),
			Active: rpcmetrics.SnapshotActive(rpcmetrics.SnapshotOptions{
				Limit:       limitActive,
				MinDuration: time.Duration(minDurationMS) * time.Millisecond,
				Service:     service,
				Phase:       phase,
			}),
		},
		Summary: summary,
	}
	ctx.Success(resp)
}

func verifyAdminAuthToken(ctx *webcontext.Context) bool {
	// Strict check: only accept valid Bearer token from .gittar/auth_token
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		ctx.AbortWithStatus(http.StatusUnauthorized, errors.New("unauthorized: no admin permission"))
		return false
	}
	bearer := strings.TrimPrefix(auth, "Bearer ")
	if !admin_token.ValidateAdminAuthToken(bearer) {
		ctx.AbortWithStatus(http.StatusForbidden, errors.New("forbidden: invalid admin auth token"))
		return false
	}
	return true
}
