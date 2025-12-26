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

package rpcmetrics

import (
	"sort"
	"time"
)

type ActiveTask struct {
	ID      string `json:"id"`
	Service string `json:"service"`
	Phase   string `json:"rpc_phase"`

	Command   string `json:"command"`
	Cmd       string `json:"phase,omitempty"`
	CmdParams string `json:"phase_params,omitempty"`

	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`

	RepoID    int64  `json:"repo_id,omitempty"`
	RepoPath  string `json:"repo_path,omitempty"`
	OrgID     int64  `json:"org_id,omitempty"`
	OrgName   string `json:"org,omitempty"`
	ProjectID int64  `json:"project_id,omitempty"`
	Project   string `json:"project,omitempty"`
	AppID     int64  `json:"app_id,omitempty"`
	App       string `json:"app,omitempty"`

	UserID    string `json:"user_id,omitempty"`
	RemoteIP  string `json:"remote_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`

	GitProtocol string    `json:"git_protocol,omitempty"`
	StartTime   time.Time `json:"start_ts"`
	DurationMS  int64     `json:"duration_ms"`
}

type ActiveSnapshot struct {
	Now time.Time `json:"now"`

	Total     int            `json:"total"`
	ByService map[string]int `json:"by_service"`
	ByPhase   map[string]int `json:"by_phase"`
	Tasks     []ActiveTask   `json:"tasks"`
}

type SnapshotOptions struct {
	Limit       int
	MinDuration time.Duration
	Service     string
	Phase       string
}

func SnapshotActive(opts SnapshotOptions) ActiveSnapshot {
	now := time.Now()

	events := make([]Event, 0, 16)
	tracker.mu.RLock()
	for _, e := range tracker.active {
		events = append(events, e)
	}
	tracker.mu.RUnlock()

	s := ActiveSnapshot{
		Now:       now,
		Total:     len(events),
		ByService: map[string]int{},
		ByPhase:   map[string]int{},
	}

	for _, e := range events {
		s.ByService[e.Service]++
		s.ByPhase[e.Phase]++

		if opts.Service != "" && e.Service != opts.Service {
			continue
		}
		if opts.Phase != "" && e.Phase != opts.Phase {
			continue
		}

		dur := now.Sub(e.Timestamp)
		if dur < opts.MinDuration {
			continue
		}

		cmd := "git " + e.Service + " --stateless-rpc"
		if e.Phase == "advertise" {
			cmd += " --advertise-refs"
		}

		s.Tasks = append(s.Tasks, ActiveTask{
			ID:          e.ID,
			Service:     e.Service,
			Phase:       e.Phase,
			Command:     cmd,
			Cmd:         e.Cmd,
			CmdParams:   e.CmdParams,
			Method:      e.Method,
			Path:        e.Path,
			RepoID:      e.RepoID,
			RepoPath:    e.RepoPath,
			OrgID:       e.OrgID,
			OrgName:     e.OrgName,
			ProjectID:   e.ProjectID,
			Project:     e.Project,
			AppID:       e.AppID,
			App:         e.App,
			UserID:      e.UserID,
			RemoteIP:    e.RemoteIP,
			UserAgent:   e.UserAgent,
			GitProtocol: e.GitProtocol,
			StartTime:   e.Timestamp,
			DurationMS:  dur.Milliseconds(),
		})
	}

	sort.Slice(s.Tasks, func(i, j int) bool {
		return s.Tasks[i].DurationMS > s.Tasks[j].DurationMS
	})

	if opts.Limit > 0 && len(s.Tasks) > opts.Limit {
		s.Tasks = s.Tasks[:opts.Limit]
	}

	return s
}
