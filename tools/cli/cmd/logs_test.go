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

package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
	"github.com/erda-project/erda/tools/cli/utils"
)

func TestFetchLogsForTaskRetriesWhenInitialQueryIsEmpty(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	origSleepForRetry := sleepForRetry
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
		sleepForRetry = origSleepForRetry
	})

	calls := 0
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls++
		if calls == 1 {
			return &apistructs.DashboardSpotLogData{Lines: []apistructs.DashboardSpotLogLine{}}, nil
		}
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{{Content: "after retry", Stream: "stdout", TimeStamp: "2026-04-14T10:00:00Z"}},
		}, nil
	}
	sleepForRetry = func(time.Duration) {}

	lines, err := fetchLogsForTask(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build"}, "stdout", 200, taskLogFetchMode{
		RetryEmpty: true,
	})
	if err != nil {
		t.Fatalf("fetchLogsForTask() error = %v", err)
	}
	if len(lines) != 1 || lines[0].Content != "after retry" {
		t.Fatalf("fetchLogsForTask() lines = %#v, want retry line", lines)
	}
}

func TestPipelineLogsTaskIDWithTailFetchesLatestLines(t *testing.T) {
	origLoadPipelineDetail := loadPipelineDetail
	origSelectPipelineTasks := selectPipelineTasks
	origFetchTaskLogData := fetchTaskLogData
	origLogsStdout := logsStdout
	t.Cleanup(func() {
		loadPipelineDetail = origLoadPipelineDetail
		selectPipelineTasks = origSelectPipelineTasks
		fetchTaskLogData = origFetchTaskLogData
		logsStdout = origLogsStdout
	})

	taskBegin := int64(1776307863000000000)
	task := common.PipelineTaskRef{
		ID:        101,
		Name:      "java-demo",
		TimeBegin: taskBegin,
	}

	loadPipelineDetail = func(*command.Context, uint64) (pipelinepb.PipelineDetailDTO, error) {
		return pipelinepb.PipelineDetailDTO{
			ID:          1000,
			OrgName:     "erda",
			ClusterName: "erda-cloud",
		}, nil
	}
	selectPipelineTasks = func(pipelinepb.PipelineDetailDTO, common.SelectTasksOptions) ([]common.PipelineTaskRef, error) {
		return []common.PipelineTaskRef{task}, nil
	}

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		if opts.Start == 0 && opts.Count == -2 {
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "tail-1", Stream: "stdout", TimeStamp: "1776307864000000000", Offset: 1},
					{Content: "tail-2", Stream: "stdout", TimeStamp: "1776307865000000000", Offset: 2},
				},
			}, nil
		}
		return &apistructs.DashboardSpotLogData{Lines: nil}, nil
	}

	var out bytes.Buffer
	logsStdout = &out

	err := PipelineLogs(&command.Context{}, 1000, 101, "", false, false, false, false, 2, "stdout", false, false)
	if err != nil {
		t.Fatalf("PipelineLogs() error = %v", err)
	}

	if len(calls) == 0 {
		t.Fatal("fetchTaskLogData() was not called")
	}
	if calls[0].Start != 0 || calls[0].Count != -2 {
		t.Fatalf("first query = %#v, want start=0 count=-2", calls[0])
	}

	got := out.String()
	want := "tail-1\ntail-2\n"
	if got != want {
		t.Fatalf("PipelineLogs() output = %q, want %q", got, want)
	}
}

func TestPipelineLogsAllFetchesEachTaskFromTaskStart(t *testing.T) {
	origLoadPipelineDetail := loadPipelineDetail
	origSelectPipelineTasks := selectPipelineTasks
	origFetchTaskLogData := fetchTaskLogData
	origLogsStdout := logsStdout
	t.Cleanup(func() {
		loadPipelineDetail = origLoadPipelineDetail
		selectPipelineTasks = origSelectPipelineTasks
		fetchTaskLogData = origFetchTaskLogData
		logsStdout = origLogsStdout
	})

	buildBegin := int64(1776307863000000000)
	testBegin := int64(1776307864000000000)
	tasks := []common.PipelineTaskRef{
		{ID: 101, Name: "build", TimeBegin: buildBegin},
		{ID: 102, Name: "test", TimeBegin: testBegin},
	}

	loadPipelineDetail = func(*command.Context, uint64) (pipelinepb.PipelineDetailDTO, error) {
		return pipelinepb.PipelineDetailDTO{
			ID:          1000,
			OrgName:     "erda",
			ClusterName: "erda-cloud",
		}, nil
	}
	selectPipelineTasks = func(pipelinepb.PipelineDetailDTO, common.SelectTasksOptions) ([]common.PipelineTaskRef, error) {
		return tasks, nil
	}

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		switch {
		case opts.TaskID == 101 && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "build-first", Stream: "stdout", TimeStamp: "1776307863000000001", Offset: 1},
				},
			}, nil
		case opts.TaskID == 101 && opts.Start == buildBegin:
			return &apistructs.DashboardSpotLogData{Lines: nil}, nil
		case opts.TaskID == 102 && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "test-first", Stream: "stdout", TimeStamp: "1776307864000000001", Offset: 1},
				},
			}, nil
		case opts.TaskID == 102 && opts.Start == testBegin:
			return &apistructs.DashboardSpotLogData{Lines: nil}, nil
		default:
			return &apistructs.DashboardSpotLogData{Lines: nil}, nil
		}
	}

	var out bytes.Buffer
	logsStdout = &out

	err := PipelineLogs(&command.Context{}, 1000, 0, "", true, false, false, false, 200, "stdout", false, false)
	if err != nil {
		t.Fatalf("PipelineLogs() error = %v", err)
	}

	if len(calls) < 2 {
		t.Fatalf("fetchTaskLogData() calls = %d, want at least 2", len(calls))
	}
	if calls[0].TaskID != 101 || calls[0].Start != 0 || calls[0].Count != 700 {
		t.Fatalf("first query = %#v, want taskID=101 start=0 count=700", calls[0])
	}
	if calls[1].TaskID != 102 || calls[1].Start != 0 || calls[1].Count != 700 {
		t.Fatalf("second query = %#v, want taskID=102 start=0 count=700", calls[1])
	}

	got := out.String()
	want := "[build#101] build-first\n[test#102] test-first\n"
	if got != want {
		t.Fatalf("PipelineLogs() output = %q, want %q", got, want)
	}
}

func TestPipelineLogsAllWithTailTrimsAggregatedOutput(t *testing.T) {
	origLoadPipelineDetail := loadPipelineDetail
	origSelectPipelineTasks := selectPipelineTasks
	origFetchTaskLogData := fetchTaskLogData
	origLogsStdout := logsStdout
	t.Cleanup(func() {
		loadPipelineDetail = origLoadPipelineDetail
		selectPipelineTasks = origSelectPipelineTasks
		fetchTaskLogData = origFetchTaskLogData
		logsStdout = origLogsStdout
	})

	taskBegin := int64(1776307863000000000)
	tasks := []common.PipelineTaskRef{
		{ID: 101, Name: "java-demo", TimeBegin: taskBegin},
		{ID: 102, Name: "git-checkout", TimeBegin: taskBegin + 100},
	}

	loadPipelineDetail = func(*command.Context, uint64) (pipelinepb.PipelineDetailDTO, error) {
		return pipelinepb.PipelineDetailDTO{
			ID:          1000,
			OrgName:     "erda",
			ClusterName: "erda-cloud",
		}, nil
	}
	selectPipelineTasks = func(pipelinepb.PipelineDetailDTO, common.SelectTasksOptions) ([]common.PipelineTaskRef, error) {
		return tasks, nil
	}

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		switch {
		case opts.TaskID == 101 && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "a1", Stream: "stdout", TimeStamp: "1776307863000000001", Offset: 1},
					{Content: "a2", Stream: "stdout", TimeStamp: "1776307863000000002", Offset: 2},
				},
			}, nil
		case opts.TaskID == 102 && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "b1", Stream: "stdout", TimeStamp: "1776307863000000003", Offset: 1},
					{Content: "b2", Stream: "stdout", TimeStamp: "1776307863000000004", Offset: 2},
				},
			}, nil
		}
		return &apistructs.DashboardSpotLogData{Lines: nil}, nil
	}

	var out bytes.Buffer
	logsStdout = &out

	err := PipelineLogs(&command.Context{}, 1000, 0, "", true, false, false, false, 1, "stdout", false, false)
	if err != nil {
		t.Fatalf("PipelineLogs() error = %v", err)
	}

	if calls[0].Count != 700 || calls[1].Count != 700 {
		t.Fatalf("query counts = %#v, want both 700", calls)
	}
	want := "[git-checkout#102] b2\n"
	if got := out.String(); got != want {
		t.Fatalf("PipelineLogs() output = %q, want %q", got, want)
	}
}

func TestFetchLogsForTaskFallsBackToForwardWindowFromTaskBegin(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	origSleepForRetry := sleepForRetry
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
		sleepForRetry = origSleepForRetry
	})

	taskBegin := int64(1776157455467386849)
	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		if len(calls) == 1 {
			return &apistructs.DashboardSpotLogData{Lines: []apistructs.DashboardSpotLogLine{}}, nil
		}
		return &apistructs.DashboardSpotLogData{
			Lines: []apistructs.DashboardSpotLogLine{{Content: "from forward window", Stream: "stdout", TimeStamp: "1776157455467388802"}},
		}, nil
	}
	sleepForRetry = func(time.Duration) {}

	lines, err := fetchLogsForTask(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build", TimeBegin: taskBegin}, "stdout", 200, taskLogFetchMode{
		RetryEmpty:      true,
		ForwardFallback: true,
	})
	if err != nil {
		t.Fatalf("fetchLogsForTask() error = %v", err)
	}
	if len(lines) != 1 || lines[0].Content != "from forward window" {
		t.Fatalf("fetchLogsForTask() lines = %#v, want forward window line", lines)
	}
	if len(calls) != 2 {
		t.Fatalf("fetchTaskLogData() calls = %d, want 2", len(calls))
	}
	if calls[0].Start != 0 || calls[0].Count != 0 {
		t.Fatalf("initial query = %#v, want default tail query", calls[0])
	}
	if calls[1].Start != taskBegin {
		t.Fatalf("fallback start = %d, want %d", calls[1].Start, taskBegin)
	}
	if calls[1].Count != 200 {
		t.Fatalf("fallback count = %d, want 200", calls[1].Count)
	}
}

func TestFetchLogsForTaskAllowsForwardWindowFallbackWithoutRetry(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	origSleepForRetry := sleepForRetry
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
		sleepForRetry = origSleepForRetry
	})

	taskBegin := int64(1776157455467386849)
	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		return &apistructs.DashboardSpotLogData{Lines: []apistructs.DashboardSpotLogLine{}}, nil
	}
	sleepForRetry = func(time.Duration) {
		t.Fatal("watch-style forward fallback should not sleep inside fetchTaskLogLines")
	}

	lines, err := fetchLogsForTask(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build", TimeBegin: taskBegin}, "stdout", 200, taskLogFetchMode{
		RetryEmpty:      false,
		ForwardFallback: true,
	})
	if err != nil {
		t.Fatalf("fetchLogsForTask() error = %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("fetchLogsForTask() lines len = %d, want 0", len(lines))
	}
	if len(calls) != 2 {
		t.Fatalf("fetchTaskLogData() calls = %d, want tail + forward fallback", len(calls))
	}
	if calls[1].Start != taskBegin || calls[1].Count != 200 {
		t.Fatalf("fallback query = %#v, want start %d and count 200", calls[1], taskBegin)
	}
}

func TestFetchTaskLogLinesFromCursorFetchesMultipleForwardBatches(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
	})

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		switch len(calls) {
		case 1:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "dup-100", Stream: "stdout", TimeStamp: "100", Offset: 1},
					{Content: "line-101", Stream: "stdout", TimeStamp: "101", Offset: 2},
				},
			}, nil
		case 2:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "line-101", Stream: "stdout", TimeStamp: "101", Offset: 2},
					{Content: "line-102", Stream: "stdout", TimeStamp: "102", Offset: 3},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "line-102", Stream: "stdout", TimeStamp: "102", Offset: 3},
				},
			}, nil
		}
	}

	cursor := &taskLogCursor{
		Initialized:   true,
		LastTimestamp: 100,
		seenAtTimestamp: map[string]struct{}{
			logCursorLineKey(apistructs.DashboardSpotLogLine{Content: "dup-100", Stream: "stdout", Offset: 1}): {},
		},
	}

	lines, err := fetchTaskLogLinesFromCursor(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build"}, "stdout", 2, cursor)
	if err != nil {
		t.Fatalf("fetchTaskLogLinesFromCursor() error = %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("fetchTaskLogLinesFromCursor() len = %d, want 2", len(lines))
	}
	if lines[0].Content != "line-101" || lines[1].Content != "line-102" {
		t.Fatalf("fetchTaskLogLinesFromCursor() lines = %#v, want line-101 and line-102", lines)
	}
	if len(calls) != 3 {
		t.Fatalf("fetchTaskLogData() calls = %d, want 3", len(calls))
	}
	if calls[0].Start != 99 || calls[0].Count != 2 {
		t.Fatalf("first query = %#v, want start 99 count 2", calls[0])
	}
	if calls[1].Start != 100 || calls[1].Count != 2 {
		t.Fatalf("second query = %#v, want start 100 count 2", calls[1])
	}
	if calls[2].Start != 101 || calls[2].Count != 2 {
		t.Fatalf("third query = %#v, want start 101 count 2", calls[2])
	}
}

func TestFetchTaskLogLinesFromCursorKeepsBoundaryLinesWithSameTimestamp(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
	})

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)
		switch len(calls) {
		case 1:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "first-at-ts", Stream: "stdout", TimeStamp: "100", Offset: 1},
					{Content: "boundary-at-ts", Stream: "stdout", TimeStamp: "100", Offset: 2},
				},
			}, nil
		case 2:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "first-at-ts", Stream: "stdout", TimeStamp: "100", Offset: 1},
					{Content: "boundary-at-ts", Stream: "stdout", TimeStamp: "100", Offset: 2},
					{Content: "third-at-same-ts", Stream: "stdout", TimeStamp: "100", Offset: 3},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{Lines: nil}, nil
		}
	}

	cursor := &taskLogCursor{
		Initialized:     true,
		LastTimestamp:   100,
		seenAtTimestamp: make(map[string]struct{}),
	}

	lines, err := fetchTaskLogLinesFromCursor(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build"}, "stdout", 2, cursor)
	if err != nil {
		t.Fatalf("fetchTaskLogLinesFromCursor() error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("fetchTaskLogLinesFromCursor() len = %d, want 3", len(lines))
	}
	if lines[2].Content != "third-at-same-ts" {
		t.Fatalf("fetchTaskLogLinesFromCursor() lines = %#v, want boundary sibling retained", lines)
	}
	if calls[0].Start != 99 {
		t.Fatalf("first query start = %d, want 99", calls[0].Start)
	}
	if calls[1].Start != 99 {
		t.Fatalf("second query start = %d, want 99", calls[1].Start)
	}
}

func TestFetchLogsForTaskSkipsRetryWhenDisabled(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
	})

	calls := 0
	fetchTaskLogData = func(_ *command.Context, _ common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls++
		return &apistructs.DashboardSpotLogData{Lines: []apistructs.DashboardSpotLogLine{}}, nil
	}

	lines, err := fetchLogsForTask(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build"}, "stdout", 200, taskLogFetchMode{})
	if err != nil {
		t.Fatalf("fetchLogsForTask() error = %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("fetchLogsForTask() lines len = %d, want 0", len(lines))
	}
	if calls != 1 {
		t.Fatalf("fetchLogsForTask() calls = %d, want 1", calls)
	}
}

func TestValidateLogOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    logOptions
		wantErr bool
	}{
		{
			name: "single task raw is allowed",
			opts: logOptions{pipelineID: 1, taskID: 2, raw: true},
		},
		{
			name:    "conflicting multi selectors are rejected",
			opts:    logOptions{pipelineID: 1, all: true, failed: true},
			wantErr: true,
		},
		{
			name:    "raw with multi task selector is rejected",
			opts:    logOptions{pipelineID: 1, all: true, raw: true},
			wantErr: true,
		},
		{
			name:    "invalid stream is rejected",
			opts:    logOptions{pipelineID: 1, stream: "weird"},
			wantErr: true,
		},
		{
			name: "pipeline id can be resolved later",
			opts: logOptions{all: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLogOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolvePipelineIDForLogsUsesLatestWorkspacePipeline(t *testing.T) {
	origGetWorkspaceBranch := getWorkspaceBranch
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origGetLatestPipelineID := getLatestPipelineID
	origStdout := os.Stdout
	t.Cleanup(func() {
		getWorkspaceBranch = origGetWorkspaceBranch
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		getLatestPipelineID = origGetLatestPipelineID
		os.Stdout = origStdout
	})

	getWorkspaceBranch = func(string) (string, error) {
		return "master", nil
	}
	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	getLatestPipelineID = func(*command.Context, uint64, string) (uint64, error) {
		return 4001, nil
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	got, err := resolvePipelineIDForLogs(&command.Context{}, 0)
	_ = w.Close()
	if err != nil {
		t.Fatalf("resolvePipelineIDForLogs() error = %v", err)
	}
	if got != 4001 {
		t.Fatalf("resolvePipelineIDForLogs() = %d, want 4001", got)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("io.ReadAll() error = %v", err)
	}
	if !bytes.Contains(out, []byte("using latest pipeline on branch master: 4001")) {
		t.Fatalf("resolvePipelineIDForLogs() output = %q, want latest pipeline hint", string(out))
	}
}

func TestPipelineLogsResolvesMissingPipelineID(t *testing.T) {
	origLoadPipelineDetail := loadPipelineDetail
	origGetWorkspaceBranch := getWorkspaceBranch
	origGetWorkspaceInfo := getWorkspaceInfo
	origGetOrgDetail := getOrgDetail
	origResolveWorkspaceApplication := resolveWorkspaceApplication
	origGetLatestPipelineID := getLatestPipelineID
	origLogsStdout := logsStdout
	origIsTTYOutput := isTTYOutput
	t.Cleanup(func() {
		loadPipelineDetail = origLoadPipelineDetail
		getWorkspaceBranch = origGetWorkspaceBranch
		getWorkspaceInfo = origGetWorkspaceInfo
		getOrgDetail = origGetOrgDetail
		resolveWorkspaceApplication = origResolveWorkspaceApplication
		getLatestPipelineID = origGetLatestPipelineID
		logsStdout = origLogsStdout
		isTTYOutput = origIsTTYOutput
	})

	getWorkspaceBranch = func(string) (string, error) {
		return "master", nil
	}
	getWorkspaceInfo = func(string, string) (utils.GitterURLInfo, error) {
		return utils.GitterURLInfo{
			OrganizationURLInfo: utils.OrganizationURLInfo{Org: "erda"},
			Project:             "demo-project",
			Application:         "demo-app",
		}, nil
	}
	getOrgDetail = func(*command.Context, string) (apistructs.OrgDTO, error) {
		return apistructs.OrgDTO{ID: 1001}, nil
	}
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string) (uint64, int64, error) {
		return 2001, 3001, nil
	}
	getLatestPipelineID = func(*command.Context, uint64, string) (uint64, error) {
		return 4001, nil
	}

	calledWith := uint64(0)
	loadPipelineDetail = func(_ *command.Context, pipelineID uint64) (pipelinepb.PipelineDetailDTO, error) {
		calledWith = pipelineID
		return pipelinepb.PipelineDetailDTO{}, nil
	}
	logsStdout = &bytes.Buffer{}
	isTTYOutput = func(io.Writer) bool { return false }

	err := PipelineLogs(&command.Context{}, 0, 0, "", true, false, false, false, 200, "stdout", false, false)
	if err == nil {
		t.Fatal("PipelineLogs() error = nil, want error from empty pipeline")
	}
	if calledWith != 4001 {
		t.Fatalf("PipelineLogs() loaded pipelineID = %d, want 4001", calledWith)
	}
}

func TestNormalizeLogStream(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		hasErr bool
	}{
		{name: "empty defaults to stdout", input: "", want: "stdout"},
		{name: "stdout preserved", input: "stdout", want: "stdout"},
		{name: "stderr preserved", input: "stderr", want: "stderr"},
		{name: "all preserved", input: "all", want: "all"},
		{name: "invalid rejected", input: "invalid", hasErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeLogStream(tt.input)
			if (err != nil) != tt.hasErr {
				t.Fatalf("normalizeLogStream() error = %v, hasErr %v", err, tt.hasErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeLogStream() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLogSelectionModePrecedence(t *testing.T) {
	tests := []struct {
		name string
		opts logOptions
		want logSelectionMode
	}{
		{
			name: "task id wins",
			opts: logOptions{taskID: 2, taskName: "build", all: true, failed: true, running: true},
			want: logSelectionModeTaskID,
		},
		{
			name: "task name over group selectors",
			opts: logOptions{taskName: "build", all: true, failed: true, running: true},
			want: logSelectionModeTaskName,
		},
		{
			name: "failed over running and all",
			opts: logOptions{all: true, failed: true, running: true},
			want: logSelectionModeFailed,
		},
		{
			name: "running over all",
			opts: logOptions{all: true, running: true},
			want: logSelectionModeRunning,
		},
		{
			name: "all over auto",
			opts: logOptions{all: true},
			want: logSelectionModeAll,
		},
		{
			name: "default auto",
			opts: logOptions{},
			want: logSelectionModeAuto,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opts.selectionMode(); got != tt.want {
				t.Fatalf("selectionMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteLogOutputSingleTaskText(t *testing.T) {
	task := common.PipelineTaskRef{ID: 101, Name: "build"}
	logs := map[uint64][]apistructs.DashboardSpotLogLine{
		101: {
			{Content: "first line", TimeStamp: "2026-04-14T10:00:00Z"},
			{Content: "second line", TimeStamp: "2026-04-14T10:00:01Z"},
		},
	}

	var buf bytes.Buffer
	err := writeLogOutput(&buf, 1000, []common.PipelineTaskRef{task}, logs, logOptions{})
	if err != nil {
		t.Fatalf("writeLogOutput() error = %v", err)
	}

	want := "first line\nsecond line\n"
	if buf.String() != want {
		t.Fatalf("writeLogOutput() = %q, want %q", buf.String(), want)
	}
}

func TestWriteLogOutputSingleTaskJSON(t *testing.T) {
	task := common.PipelineTaskRef{ID: 101, Name: "build"}
	logs := map[uint64][]apistructs.DashboardSpotLogLine{
		101: {
			{Content: "hello", Stream: "stdout", TimeStamp: "2026-04-14T10:00:00Z"},
		},
	}

	var buf bytes.Buffer
	err := writeLogOutput(&buf, 1000, []common.PipelineTaskRef{task}, logs, logOptions{jsonOutput: true})
	if err != nil {
		t.Fatalf("writeLogOutput() error = %v", err)
	}

	want := "{\"pipelineID\":1000,\"taskID\":101,\"taskName\":\"build\",\"stream\":\"stdout\",\"timestamp\":\"2026-04-14T10:00:00Z\",\"content\":\"hello\"}\n"
	if buf.String() != want {
		t.Fatalf("writeLogOutput() = %q, want %q", buf.String(), want)
	}
}

func TestWriteLogOutputMergedText(t *testing.T) {
	tasks := []common.PipelineTaskRef{
		{ID: 101, Name: "build"},
		{ID: 102, Name: "test"},
	}
	logs := map[uint64][]apistructs.DashboardSpotLogLine{
		101: {
			{Content: "compile", TimeStamp: "2026-04-14T10:00:01Z"},
		},
		102: {
			{Content: "run case", TimeStamp: "2026-04-14T10:00:02Z"},
		},
	}

	var buf bytes.Buffer
	err := writeLogOutput(&buf, 1000, tasks, logs, logOptions{all: true})
	if err != nil {
		t.Fatalf("writeLogOutput() error = %v", err)
	}

	want := "[build#101] compile\n[test#102] run case\n"
	if buf.String() != want {
		t.Fatalf("writeLogOutput() = %q, want %q", buf.String(), want)
	}
}

func TestTaskLogCursorFilterAndAdvance(t *testing.T) {
	cursor := &taskLogCursor{}
	lines := []apistructs.DashboardSpotLogLine{
		{Content: "compile", Stream: "stdout", TimeStamp: "2026-04-14T10:00:01Z", Offset: 1},
		{Content: "compile", Stream: "stdout", TimeStamp: "2026-04-14T10:00:01Z", Offset: 1},
		{Content: "test", Stream: "stdout", TimeStamp: "2026-04-14T10:00:02Z", Offset: 2},
	}

	first := cursor.filterAndAdvance(lines)
	if len(first) != 2 {
		t.Fatalf("first filter len = %d, want 2", len(first))
	}

	second := cursor.filterAndAdvance(lines)
	if len(second) != 0 {
		t.Fatalf("second filter len = %d, want 0", len(second))
	}
}

func TestDefaultSelectedTasks(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC).UnixNano()
	tasks := []common.PipelineTaskRef{
		{ID: 101, Name: "build", Status: apistructs.PipelineStatusSuccess.String(), TimeBegin: now},
		{ID: 102, Name: "test", Status: apistructs.PipelineStatusFailed.String(), TimeBegin: now + int64(time.Minute)},
		{ID: 103, Name: "release", Status: apistructs.PipelineStatusRunning.String(), TimeBegin: now + int64(2*time.Minute)},
	}

	got := defaultSelectedTasks(tasks, logSelectionModeAuto, true)
	if len(got) != 1 || got[0].ID != 103 {
		t.Fatalf("defaultSelectedTasks() interactive = %#v, want running task 103", got)
	}

	got = defaultSelectedTasks(tasks, logSelectionModeAuto, false)
	if len(got) != 3 {
		t.Fatalf("defaultSelectedTasks() non-interactive len = %d, want 3", len(got))
	}
}
