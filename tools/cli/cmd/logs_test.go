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

func TestFetchIncrementalLogLinesPagesBeyondTailWindow(t *testing.T) {
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
					{Content: "line-1", Stream: "stdout", TimeStamp: "1001"},
					{Content: "line-2", Stream: "stdout", TimeStamp: "1002"},
				},
			}, nil
		case 2:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Content: "line-3", Stream: "stdout", TimeStamp: "1003"},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{}, nil
		}
	}

	queryOpts := common.TaskLogRequestOptions{
		PipelineID: 1000,
		TaskID:     101,
		OrgName:    "erda",
		Stream:     "stdout",
		Start:      1000,
		Count:      2,
		Tail:       2,
	}
	lines, err := common.FetchIncrementalLogLines(queryOpts.Start, common.DefaultLogPageSize(queryOpts.Tail, queryOpts.Count), func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
		queryOpts.Start = start
		queryOpts.Count = count
		data, err := fetchTaskLogData(&command.Context{}, queryOpts)
		if err != nil {
			return nil, err
		}
		return data.Lines, nil
	})
	if err != nil {
		t.Fatalf("fetchIncrementalTaskLogLines() error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("fetchIncrementalTaskLogLines() len = %d, want 3", len(lines))
	}
	if len(calls) != 2 {
		t.Fatalf("fetchIncrementalTaskLogLines() calls = %d, want 2", len(calls))
	}
	if calls[0].Start != 999 || calls[0].Count != 2 {
		t.Fatalf("first call = %#v, want start 999 count 2", calls[0])
	}
	if calls[1].Start != 1001 || calls[1].Count != 2 {
		t.Fatalf("second call = %#v, want start 1001 count 2", calls[1])
	}
}

func TestFetchWatchLogLinesUsesPageFetcher(t *testing.T) {
	initialCalls := 0
	var pageCalls []struct {
		start int64
		count int64
	}

	lines, cursor, err := common.FetchWatchLogLines(common.LogCursor{}, 2, func() ([]apistructs.DashboardSpotLogLine, error) {
		initialCalls++
		return []apistructs.DashboardSpotLogLine{
			{TimeStamp: "1000", Content: "seed"},
		}, nil
	}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
		pageCalls = append(pageCalls, struct {
			start int64
			count int64
		}{start: start, count: count})
		switch len(pageCalls) {
		case 1:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1001", Content: "line-1"},
				{TimeStamp: "1002", Content: "line-2"},
			}, nil
		case 2:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1003", Content: "line-3"},
			}, nil
		default:
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("fetchWatchLogLines(initial) error = %v", err)
	}
	if initialCalls != 1 {
		t.Fatalf("initial calls = %d, want 1", initialCalls)
	}
	if len(lines) != 1 || cursor.Start != 1000 {
		t.Fatalf("initial result = (%d lines, cursor=%d), want 1 line and cursor 1000", len(lines), cursor.Start)
	}

	lines, cursor, err = common.FetchWatchLogLines(cursor, 2, func() ([]apistructs.DashboardSpotLogLine, error) {
		t.Fatal("fetchInitial should not be called after cursor is set")
		return nil, nil
	}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
		pageCalls = append(pageCalls, struct {
			start int64
			count int64
		}{start: start, count: count})
		switch len(pageCalls) {
		case 1:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1001", Content: "line-1"},
				{TimeStamp: "1002", Content: "line-2"},
			}, nil
		case 2:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1003", Content: "line-3"},
			}, nil
		default:
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("fetchWatchLogLines(incremental) error = %v", err)
	}
	if len(pageCalls) != 2 {
		t.Fatalf("page fetch calls = %d, want 2", len(pageCalls))
	}
	if pageCalls[0].start != 999 || pageCalls[0].count != 2 {
		t.Fatalf("first page call = %#v, want start 999 count 2", pageCalls[0])
	}
	if pageCalls[1].start != 1001 || pageCalls[1].count != 2 {
		t.Fatalf("second page call = %#v, want start 1001 count 2", pageCalls[1])
	}
	if len(lines) != 3 || cursor.Start != 1003 {
		t.Fatalf("incremental result = (%d lines, cursor=%d), want 3 lines and cursor 1003", len(lines), cursor.Start)
	}
}

func TestFetchWatchLogLinesExpandsPageAtTimestampBoundary(t *testing.T) {
	var calls []struct {
		start int64
		count int64
	}

	lines, cursor, err := common.FetchWatchLogLines(common.LogCursor{Start: 1000}, 2, func() ([]apistructs.DashboardSpotLogLine, error) {
		t.Fatal("fetchInitial should not be called when cursor is set")
		return nil, nil
	}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
		calls = append(calls, struct {
			start int64
			count int64
		}{start: start, count: count})
		switch len(calls) {
		case 1:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1000", Content: "line-1"},
				{TimeStamp: "1000", Content: "line-2"},
			}, nil
		case 2:
			return []apistructs.DashboardSpotLogLine{
				{TimeStamp: "1000", Content: "line-1"},
				{TimeStamp: "1000", Content: "line-2"},
				{TimeStamp: "1000", Content: "line-3"},
				{TimeStamp: "1001", Content: "line-4"},
			}, nil
		default:
			return nil, nil
		}
	})
	if err != nil {
		t.Fatalf("FetchWatchLogLines() error = %v", err)
	}
	if len(calls) != 3 {
		t.Fatalf("page fetch calls = %d, want 3", len(calls))
	}
	if calls[0].start != 999 || calls[0].count != 2 {
		t.Fatalf("first page call = %#v, want start 999 count 2", calls[0])
	}
	if calls[1].start != 999 || calls[1].count != 4 {
		t.Fatalf("second page call = %#v, want start 999 count 4", calls[1])
	}
	if calls[2].start != 1000 || calls[2].count != 2 {
		t.Fatalf("third page call = %#v, want start 1000 count 2", calls[2])
	}
	if len(lines) != 4 || cursor.Start != 1001 {
		t.Fatalf("result = (%d lines, cursor=%d), want 4 lines and cursor 1001", len(lines), cursor.Start)
	}
}

func TestFetchLogsForTaskWatchAllUsesIndependentStreamCursors(t *testing.T) {
	origFetchTaskLogData := fetchTaskLogData
	t.Cleanup(func() {
		fetchTaskLogData = origFetchTaskLogData
	})

	var calls []common.TaskLogRequestOptions
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		calls = append(calls, opts)

		if opts.Stream == "stdout" && opts.Start == 1199 {
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stdout", TimeStamp: "1300", Content: "stdout-1300"},
					{Stream: "stdout", TimeStamp: "1400", Content: "stdout-1400"},
				},
			}, nil
		}
		if opts.Stream == "stdout" && opts.Start == 1399 {
			return &apistructs.DashboardSpotLogData{}, nil
		}
		if opts.Stream == "stderr" && opts.Start == 1999 {
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stderr", TimeStamp: "2100", Content: "stderr-2100"},
				},
			}, nil
		}

		return &apistructs.DashboardSpotLogData{}, nil
	}

	lines, cursor, err := fetchLogsForTaskWatch(&command.Context{}, pipelinepb.PipelineDetailDTO{
		ID:          1000,
		OrgName:     "erda",
		ClusterName: "erda-cloud",
	}, common.PipelineTaskRef{ID: 101, Name: "build"}, "all", 2, taskLogWatchCursor{
		stdout: common.LogCursor{Start: 1200},
		stderr: common.LogCursor{Start: 2000},
	})
	if err != nil {
		t.Fatalf("fetchLogsForTaskWatch() error = %v", err)
	}

	if len(lines) != 3 {
		t.Fatalf("fetchLogsForTaskWatch() lines len = %d, want 3", len(lines))
	}
	if lines[0].TimeStamp != "1300" || lines[1].TimeStamp != "1400" || lines[2].TimeStamp != "2100" {
		t.Fatalf("fetchLogsForTaskWatch() lines = %#v, want sorted merged stdout/stderr lines", lines)
	}
	if cursor.stdout.Start != 1400 {
		t.Fatalf("stdout cursor = %d, want 1400", cursor.stdout.Start)
	}
	if cursor.stderr.Start != 2100 {
		t.Fatalf("stderr cursor = %d, want 2100", cursor.stderr.Start)
	}
	if cursor.merged.Start != 0 {
		t.Fatalf("merged cursor = %d, want 0 in all-stream mode", cursor.merged.Start)
	}

	var stdoutFirstCall, stderrFirstCall common.TaskLogRequestOptions
	for _, call := range calls {
		if call.Stream == "stdout" && call.Start == 1199 {
			stdoutFirstCall = call
		}
		if call.Stream == "stderr" && call.Start == 1999 {
			stderrFirstCall = call
		}
	}
	if stdoutFirstCall.Start == 0 {
		t.Fatalf("stdout incremental call not found in %#v", calls)
	}
	if stderrFirstCall.Start == 0 {
		t.Fatalf("stderr incremental call not found in %#v", calls)
	}
}

func TestParseLogTimestamp(t *testing.T) {
	rfcTime := "2026-04-22T10:00:00Z"
	got, ok := common.ParseLogTimestamp(rfcTime)
	if !ok || got != time.Date(2026, 4, 22, 10, 0, 0, 0, time.UTC).UnixNano() {
		t.Fatalf("ParseLogTimestamp(%q) = (%d, %v)", rfcTime, got, ok)
	}

	got, ok = common.ParseLogTimestamp("1776157455467386849")
	if !ok || got != 1776157455467386849 {
		t.Fatalf("ParseLogTimestamp(nanos) = (%d, %v)", got, ok)
	}

	if _, ok := common.ParseLogTimestamp("bad"); ok {
		t.Fatal("ParseLogTimestamp(bad) should fail")
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
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string, string) (uint64, int64, error) {
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
	resolveWorkspaceApplication = func(*command.Context, uint64, string, string, string) (uint64, int64, error) {
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

func TestWatchPipelineLogsAllUsesIndependentStreamCursorsAcrossPolls(t *testing.T) {
	origLoadPipelineDetail := loadPipelineDetail
	origSelectPipelineTasks := selectPipelineTasks
	origFetchTaskLogData := fetchTaskLogData
	origSleepForWatchPoll := sleepForWatchPoll
	origLogsStdout := logsStdout
	origIsTTYOutput := isTTYOutput
	t.Cleanup(func() {
		loadPipelineDetail = origLoadPipelineDetail
		selectPipelineTasks = origSelectPipelineTasks
		fetchTaskLogData = origFetchTaskLogData
		sleepForWatchPoll = origSleepForWatchPoll
		logsStdout = origLogsStdout
		isTTYOutput = origIsTTYOutput
	})

	task := common.PipelineTaskRef{ID: 101, Name: "build", TimeBegin: 1000}
	round := 0
	loadPipelineDetail = func(_ *command.Context, pipelineID uint64) (pipelinepb.PipelineDetailDTO, error) {
		round++
		status := apistructs.PipelineStatusRunning.String()
		if round >= 2 {
			status = apistructs.PipelineStatusSuccess.String()
		}
		return pipelinepb.PipelineDetailDTO{
			ID:          pipelineID,
			OrgName:     "erda",
			ClusterName: "erda-cloud",
			Status:      status,
		}, nil
	}
	selectPipelineTasks = func(_ pipelinepb.PipelineDetailDTO, _ common.SelectTasksOptions) ([]common.PipelineTaskRef, error) {
		return []common.PipelineTaskRef{task}, nil
	}

	type fetchKey struct {
		stream string
		start  int64
	}
	callsByKey := map[fetchKey]int{}
	fetchTaskLogData = func(_ *command.Context, opts common.TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
		key := fetchKey{stream: opts.Stream, start: opts.Start}
		callsByKey[key]++

		switch {
		case opts.Stream == "stdout" && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stdout", TimeStamp: "1200", Content: "stdout-1200"},
				},
			}, nil
		case opts.Stream == "stderr" && opts.Start == 0:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stderr", TimeStamp: "2000", Content: "stderr-2000"},
				},
			}, nil
		case opts.Stream == "stdout" && opts.Start == 1199:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stdout", TimeStamp: "1300", Content: "stdout-1300"},
				},
			}, nil
		case opts.Stream == "stderr" && opts.Start == 1999:
			return &apistructs.DashboardSpotLogData{
				Lines: []apistructs.DashboardSpotLogLine{
					{Stream: "stderr", TimeStamp: "2100", Content: "stderr-2100"},
				},
			}, nil
		default:
			return &apistructs.DashboardSpotLogData{}, nil
		}
	}

	sleepCalls := 0
	sleepForWatchPoll = func(time.Duration) {
		sleepCalls++
	}

	var out bytes.Buffer
	logsStdout = &out
	isTTYOutput = func(io.Writer) bool { return false }

	err := watchPipelineLogs(&command.Context{}, 1000, logOptions{
		watch:  true,
		all:    true,
		stream: "all",
		tail:   1,
	})
	if err != nil {
		t.Fatalf("watchPipelineLogs() error = %v", err)
	}

	got := out.String()
	want := "stdout-1200\nstderr-2000\nstdout-1300\nstderr-2100\n"
	if got != want {
		t.Fatalf("watchPipelineLogs() output = %q, want %q", got, want)
	}
	if sleepCalls != 3 {
		t.Fatalf("watch poll sleeps = %d, want 3 with two idle rounds before exit", sleepCalls)
	}
	if callsByKey[fetchKey{stream: "stdout", start: 1199}] == 0 {
		t.Fatalf("stdout incremental call missing; calls = %#v", callsByKey)
	}
	if callsByKey[fetchKey{stream: "stderr", start: 1999}] == 0 {
		t.Fatalf("stderr incremental call missing; calls = %#v", callsByKey)
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

func TestFilterUnseenLogLines(t *testing.T) {
	seen := map[uint64]map[string]struct{}{}
	lines := []apistructs.DashboardSpotLogLine{
		{Content: "compile", Stream: "stdout", TimeStamp: "2026-04-14T10:00:01Z", Offset: 1},
		{Content: "compile", Stream: "stdout", TimeStamp: "2026-04-14T10:00:01Z", Offset: 1},
		{Content: "test", Stream: "stdout", TimeStamp: "2026-04-14T10:00:02Z", Offset: 2},
	}

	first := filterUnseenLogLines(101, lines, seen)
	if len(first) != 2 {
		t.Fatalf("first filter len = %d, want 2", len(first))
	}

	second := filterUnseenLogLines(101, lines, seen)
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
