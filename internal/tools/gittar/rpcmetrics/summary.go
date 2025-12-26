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
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RepoPullSummary struct {
	RepoID     int64  `json:"repo_id,omitempty"`
	RepoPath   string `json:"repo_path,omitempty"`
	Org        string `json:"org,omitempty"`
	Project    string `json:"project,omitempty"`
	App        string `json:"app,omitempty"`
	Count      int    `json:"count"`
	Depth1     int    `json:"depth_1"`
	DepthOther int    `json:"depth_other"`
	DepthNone  int    `json:"depth_none"`
}

type RepoPushSummary struct {
	RepoID   int64  `json:"repo_id,omitempty"`
	RepoPath string `json:"repo_path,omitempty"`
	Org      string `json:"org,omitempty"`
	Project  string `json:"project,omitempty"`
	App      string `json:"app,omitempty"`
	Count    int    `json:"count"`
}

type TopPullRepos struct {
	Limit   int               `json:"limit"`
	Details []RepoPullSummary `json:"details"`
}

type TopPushRepos struct {
	Limit   int               `json:"limit"`
	Details []RepoPushSummary `json:"details"`
}

type TopRepos struct {
	Pull TopPullRepos `json:"pull"`
	Push TopPushRepos `json:"push"`
}

type Summary struct {
	Date string `json:"date"`

	CompletedTotal     int            `json:"completed_total"`
	CompletedByService map[string]int `json:"completed_by_service"`
	CompletedByPhase   map[string]int `json:"completed_by_phase"`
	CompletedErrors    int            `json:"completed_errors"`

	PullsTotal  int `json:"pulls_total"`
	PushesTotal int `json:"pushes_total"`

	Top TopRepos `json:"top"`

	Error string `json:"error,omitempty"`
}

func BuildSummary(filePath string, topN int) Summary {
	summary := Summary{
		Date:               summaryDate(filePath),
		CompletedByService: map[string]int{},
		CompletedByPhase:   map[string]int{},
		Top: TopRepos{
			Pull: TopPullRepos{Limit: topN, Details: []RepoPullSummary{}},
			Push: TopPushRepos{Limit: topN, Details: []RepoPushSummary{}},
		},
	}
	if filePath == "" {
		return summary
	}

	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return summary
		}
		summary.Error = err.Error()
		return summary
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	type repoCounter struct {
		summary RepoPullSummary
	}
	pullMap := map[string]*repoCounter{}

	type repoPushCounter struct {
		summary RepoPushSummary
	}
	pushMap := map[string]*repoPushCounter{}

	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue
		}
		if e.Event != "end" {
			continue
		}

		summary.CompletedTotal++
		if e.Service != "" {
			summary.CompletedByService[e.Service]++
		}
		if e.Phase != "" {
			summary.CompletedByPhase[e.Phase]++
		}
		if e.Error != "" {
			summary.CompletedErrors++
		}

		if e.Service == "upload-pack" && e.Phase == "rpc" {
			if e.Cmd == "" || e.Cmd == "fetch" {
				summary.PullsTotal++
				key := repoKey(e)
				counter := pullMap[key]
				if counter == nil {
					counter = &repoCounter{
						summary: RepoPullSummary{
							RepoID:   e.RepoID,
							RepoPath: e.RepoPath,
							Org:      e.OrgName,
							Project:  e.Project,
							App:      e.App,
						},
					}
					pullMap[key] = counter
				}
				counter.summary.Count++
				switch parseDepthKind(e.CmdParams) {
				case depthKind1:
					counter.summary.Depth1++
				case depthKindOther:
					counter.summary.DepthOther++
				default:
					counter.summary.DepthNone++
				}
			}
		}
		if e.Service == "receive-pack" && e.Phase == "rpc" {
			summary.PushesTotal++
			key := repoKey(e)
			counter := pushMap[key]
			if counter == nil {
				counter = &repoPushCounter{
					summary: RepoPushSummary{
						RepoID:   e.RepoID,
						RepoPath: e.RepoPath,
						Org:      e.OrgName,
						Project:  e.Project,
						App:      e.App,
					},
				}
				pushMap[key] = counter
			}
			counter.summary.Count++
		}
	}
	if err := scanner.Err(); err != nil {
		summary.Error = err.Error()
	}

	if topN <= 0 {
		return summary
	}

	// Pulls
	if len(pullMap) > 0 {
		list := make([]RepoPullSummary, 0, len(pullMap))
		for _, counter := range pullMap {
			list = append(list, counter.summary)
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Count == list[j].Count {
				return list[i].RepoPath < list[j].RepoPath
			}
			return list[i].Count > list[j].Count
		})
		if len(list) > topN {
			list = list[:topN]
		}
		summary.Top.Pull.Details = list
	}

	// Pushes
	if len(pushMap) > 0 {
		list := make([]RepoPushSummary, 0, len(pushMap))
		for _, counter := range pushMap {
			list = append(list, counter.summary)
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Count == list[j].Count {
				return list[i].RepoPath < list[j].RepoPath
			}
			return list[i].Count > list[j].Count
		})
		if len(list) > topN {
			list = list[:topN]
		}
		summary.Top.Push.Details = list
	}

	return summary
}

func summaryDate(filePath string) string {
	if filePath == "" {
		return time.Now().Format("2006-01-02")
	}
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if name == "" {
		return time.Now().Format("2006-01-02")
	}
	if t, err := time.ParseInLocation("2006-01-02", name, time.Local); err == nil {
		return t.Format("2006-01-02")
	}
	return time.Now().Format("2006-01-02")
}

func repoKey(e Event) string {
	if e.RepoPath != "" {
		return e.RepoPath
	}
	if e.RepoID != 0 {
		return "id:" + strconv.FormatInt(e.RepoID, 10)
	}
	return ""
}

type depthKind int

const (
	depthKindNone depthKind = iota
	depthKind1
	depthKindOther
)

func parseDepthKind(cmdParams string) depthKind {
	params := strings.TrimSpace(cmdParams)
	if params == "" {
		return depthKindNone
	}
	parts := strings.Fields(params)
	depth := ""
	for _, part := range parts {
		if strings.HasPrefix(part, "deepen=") {
			depth = strings.TrimPrefix(part, "deepen=")
			break
		}
		if strings.HasPrefix(part, "depth=") {
			depth = strings.TrimPrefix(part, "depth=")
			break
		}
	}
	if depth == "1" {
		return depthKind1
	}
	if depth != "" || strings.Contains(params, "deepen_since=") || strings.Contains(params, "deepen_not=") {
		return depthKindOther
	}
	return depthKindNone
}
