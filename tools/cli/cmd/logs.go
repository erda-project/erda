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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/common"
)

type logSelectionMode string

const (
	logSelectionModeAuto     logSelectionMode = "auto"
	logSelectionModeAll      logSelectionMode = "all"
	logSelectionModeRunning  logSelectionMode = "running"
	logSelectionModeFailed   logSelectionMode = "failed"
	logSelectionModeTaskName logSelectionMode = "task-name"
	logSelectionModeTaskID   logSelectionMode = "task-id"
)

const (
	emptyLogFetchAttempts = 3
	emptyLogRetryInterval = time.Second
)

type logOptions struct {
	pipelineID uint64
	taskID     uint64
	taskName   string
	all        bool
	failed     bool
	running    bool
	watch      bool
	tail       int
	stream     string
	jsonOutput bool
	raw        bool
}

type taskLogFetchMode struct {
	RetryEmpty      bool
	ForwardFallback bool
}

var LOGS = command.Command{
	ParentName: "PIPELINE",
	Name:       "logs",
	ShortHelp:  "View pipeline logs",
	Example: ` $ erda-cli pipeline logs --all
  $ erda-cli pipeline logs -i <pipelineID>
  $ erda-cli pipeline logs -i <pipelineID> --task-id <taskID>
  $ erda-cli pipeline logs -i <pipelineID> --failed --watch`,
	Flags: []command.Flag{
		command.Uint64Flag{Short: "i", Name: "pipelineID", Doc: "specify pipeline id to show logs", DefaultValue: 0},
		command.Uint64Flag{Short: "", Name: "task-id", Doc: "show logs for an exact task id", DefaultValue: 0},
		command.StringFlag{Short: "", Name: "task", Doc: "show logs for a task name", DefaultValue: ""},
		command.BoolFlag{Short: "", Name: "all", Doc: "merge all task logs in the pipeline", DefaultValue: false},
		command.BoolFlag{Short: "", Name: "failed", Doc: "merge failed task logs only", DefaultValue: false},
		command.BoolFlag{Short: "", Name: "running", Doc: "merge running task logs only", DefaultValue: false},
		command.BoolFlag{Short: "w", Name: "watch", Doc: "watch for new log lines", DefaultValue: false},
		command.IntFlag{Short: "", Name: "tail", Doc: "number of recent log lines to fetch first", DefaultValue: 200},
		command.StringFlag{Short: "", Name: "stream", Doc: "log stream: stdout, stderr, or all", DefaultValue: ""},
		command.BoolFlag{Short: "", Name: "json", Doc: "output logs as JSON lines", DefaultValue: false},
		command.BoolFlag{Short: "", Name: "raw", Doc: "emit raw content only for a single task selection", DefaultValue: false},
	},
	Run: PipelineLogs,
}

var (
	loadPipelineDetail            = common.GetPipeline
	selectPipelineTasks           = common.SelectTasks
	fetchTaskLogData              = common.GetTaskLog
	logsStdout          io.Writer = os.Stdout
	isTTYOutput                   = isTTYWriter
	sleepForRetry                 = time.Sleep
	sleepForWatchPoll             = time.Sleep
)

func PipelineLogs(ctx *command.Context, pipelineID uint64, taskID uint64, taskName string, all bool, failed bool, running bool, watch bool, tail int, stream string, jsonOutput bool, raw bool) error {
	opts := logOptions{
		pipelineID: pipelineID,
		taskID:     taskID,
		taskName:   taskName,
		all:        all,
		failed:     failed,
		running:    running,
		watch:      watch,
		tail:       tail,
		stream:     stream,
		jsonOutput: jsonOutput,
		raw:        raw,
	}
	if err := validateLogOptions(opts); err != nil {
		return err
	}
	opts.stream, _ = normalizeLogStream(opts.stream)
	var err error
	opts.pipelineID, err = resolvePipelineIDForLogs(ctx, opts.pipelineID)
	if err != nil {
		return err
	}
	pipelineID = opts.pipelineID

	pipeline, err := loadPipelineDetail(ctx, pipelineID)
	if err != nil {
		return err
	}
	tasks, err := selectPipelineTasks(pipeline, common.SelectTasksOptions{
		TaskID:      opts.taskID,
		TaskName:    opts.taskName,
		All:         opts.all,
		FailedOnly:  opts.failed,
		RunningOnly: opts.running,
	})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return fmt.Errorf("no tasks matched in pipeline %d", pipelineID)
	}
	tasks = defaultSelectedTasks(tasks, opts.selectionMode(), isTTYOutput(logsStdout))

	if opts.watch {
		return watchPipelineLogs(ctx, pipelineID, opts)
	}

	logsByTask := make(map[uint64][]apistructs.DashboardSpotLogLine, len(tasks))
	for _, task := range tasks {
		lines, err := fetchLogsForTask(ctx, pipeline, task, opts.stream, opts.tail, taskLogFetchMode{
			RetryEmpty:      true,
			ForwardFallback: true,
		})
		if err != nil {
			return err
		}
		logsByTask[task.ID] = lines
	}

	return writeLogOutput(logsStdout, pipelineID, tasks, logsByTask, opts)
}

func (o logOptions) selectionMode() logSelectionMode {
	switch {
	case o.taskID > 0:
		return logSelectionModeTaskID
	case o.taskName != "":
		return logSelectionModeTaskName
	case o.failed:
		return logSelectionModeFailed
	case o.running:
		return logSelectionModeRunning
	case o.all:
		return logSelectionModeAll
	default:
		return logSelectionModeAuto
	}
}

func (o logOptions) multiTaskSelectorCount() int {
	count := 0
	if o.all {
		count++
	}
	if o.failed {
		count++
	}
	if o.running {
		count++
	}
	return count
}

func normalizeLogStream(stream string) (string, error) {
	switch stream {
	case "", "stdout":
		return "stdout", nil
	case "stderr":
		return "stderr", nil
	case "all":
		return "all", nil
	default:
		return "", fmt.Errorf("invalid stream %q, supported values: stdout, stderr, all", stream)
	}
}

func validateLogOptions(opts logOptions) error {
	if opts.multiTaskSelectorCount() > 1 {
		return errors.New("flags --all, --failed, and --running are mutually exclusive")
	}
	if _, err := normalizeLogStream(opts.stream); err != nil {
		return err
	}
	switch opts.selectionMode() {
	case logSelectionModeAll, logSelectionModeFailed, logSelectionModeRunning:
		if opts.raw {
			return errors.New("--raw is only supported for a single task selection")
		}
	}
	return nil
}

func resolvePipelineIDForLogs(ctx *command.Context, pipelineID uint64) (uint64, error) {
	if pipelineID > 0 {
		return pipelineID, nil
	}

	branch, err := getWorkspaceBranch(".")
	if err != nil {
		return 0, err
	}
	info, err := getWorkspaceInfo(".", command.Remote)
	if err != nil {
		return 0, err
	}
	org, err := getOrgDetail(ctx, info.Org)
	if err != nil {
		return 0, err
	}
	_, applicationID, err := resolveWorkspaceApplication(ctx, org.ID, info.Project, info.Application)
	if err != nil {
		return 0, err
	}
	pipelineID, err = getLatestPipelineID(ctx, uint64(applicationID), branch)
	if err != nil {
		return 0, err
	}
	ctx.Info("no pipeline id provided, using latest pipeline on branch %s: %d", branch, pipelineID)
	return pipelineID, nil
}

type logOutputRecord struct {
	PipelineID uint64 `json:"pipelineID"`
	TaskID     uint64 `json:"taskID"`
	TaskName   string `json:"taskName"`
	Stream     string `json:"stream"`
	Timestamp  string `json:"timestamp"`
	Content    string `json:"content"`
}

type mergedLogLine struct {
	task common.PipelineTaskRef
	line apistructs.DashboardSpotLogLine
}

type taskLogWatchCursor struct {
	merged common.LogCursor
	stdout common.LogCursor
	stderr common.LogCursor
}

func watchPipelineLogs(ctx *command.Context, pipelineID uint64, opts logOptions) error {
	seen := make(map[uint64]map[string]struct{})
	cursors := make(map[uint64]taskLogWatchCursor)
	idleRounds := 0

	for {
		pipeline, err := loadPipelineDetail(ctx, pipelineID)
		if err != nil {
			return err
		}
		tasks, err := selectPipelineTasks(pipeline, common.SelectTasksOptions{
			TaskID:      opts.taskID,
			TaskName:    opts.taskName,
			All:         opts.all,
			FailedOnly:  opts.failed,
			RunningOnly: opts.running,
		})
		if err != nil {
			return err
		}
		if len(tasks) == 0 {
			return fmt.Errorf("no tasks matched in pipeline %d", pipelineID)
		}
		tasks = defaultSelectedTasks(tasks, opts.selectionMode(), isTTYOutput(logsStdout))

		logsByTask := make(map[uint64][]apistructs.DashboardSpotLogLine, len(tasks))
		newLineCount := 0
		for _, task := range tasks {
			lines, nextCursor, err := fetchLogsForTaskWatch(ctx, pipeline, task, opts.stream, opts.tail, cursors[task.ID])
			if err != nil {
				return err
			}
			cursors[task.ID] = nextCursor
			unseen := filterUnseenLogLines(task.ID, lines, seen)
			newLineCount += len(unseen)
			logsByTask[task.ID] = unseen
		}

		if newLineCount > 0 {
			idleRounds = 0
			if err := writeLogOutput(logsStdout, pipelineID, tasks, logsByTask, opts); err != nil {
				return err
			}
		} else {
			idleRounds++
		}

		if apistructs.PipelineStatus(pipeline.Status).IsEndStatus() && idleRounds >= 2 {
			return nil
		}

		sleepForWatchPoll(2 * time.Second)
	}
}

func fetchLogsForTaskWatch(ctx *command.Context, pipeline pipelinepb.PipelineDetailDTO, task common.PipelineTaskRef, stream string, tail int, cursor taskLogWatchCursor) ([]apistructs.DashboardSpotLogLine, taskLogWatchCursor, error) {
	if stream == "all" {
		stdoutLines, stdoutCursor, err := common.FetchWatchLogLines(cursor.stdout, tail, func() ([]apistructs.DashboardSpotLogLine, error) {
			return fetchLogsForTask(ctx, pipeline, task, "stdout", tail, taskLogFetchMode{
				ForwardFallback: true,
			})
		}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
			return fetchTaskLogPage(ctx, pipeline, task, "stdout", tail, start, count)
		})
		if err != nil {
			return nil, cursor, err
		}

		stderrLines, stderrCursor, err := common.FetchWatchLogLines(cursor.stderr, tail, func() ([]apistructs.DashboardSpotLogLine, error) {
			return fetchLogsForTask(ctx, pipeline, task, "stderr", tail, taskLogFetchMode{
				ForwardFallback: true,
			})
		}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
			return fetchTaskLogPage(ctx, pipeline, task, "stderr", tail, start, count)
		})
		if err != nil {
			return nil, cursor, err
		}

		cursor.stdout = stdoutCursor
		cursor.stderr = stderrCursor
		return mergeAndSortTaskLogLines(stdoutLines, stderrLines), cursor, nil
	}

	lines, nextCursor, err := common.FetchWatchLogLines(cursor.merged, tail, func() ([]apistructs.DashboardSpotLogLine, error) {
		return fetchLogsForTask(ctx, pipeline, task, stream, tail, taskLogFetchMode{
			ForwardFallback: true,
		})
	}, func(start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
		return fetchTaskLogPage(ctx, pipeline, task, stream, tail, start, count)
	})
	if err != nil {
		return nil, cursor, err
	}
	cursor.merged = nextCursor
	return lines, cursor, nil
}

func mergeAndSortTaskLogLines(left []apistructs.DashboardSpotLogLine, right []apistructs.DashboardSpotLogLine) []apistructs.DashboardSpotLogLine {
	lines := append([]apistructs.DashboardSpotLogLine{}, left...)
	lines = append(lines, right...)
	sort.SliceStable(lines, func(i, j int) bool {
		return lines[i].TimeStamp < lines[j].TimeStamp
	})
	return lines
}

func fetchLogsForTask(ctx *command.Context, pipeline pipelinepb.PipelineDetailDTO, task common.PipelineTaskRef, stream string, tail int, mode taskLogFetchMode) ([]apistructs.DashboardSpotLogLine, error) {
	if stream == "all" {
		stdoutLines, err := fetchTaskLogLines(ctx, common.TaskLogRequestOptions{
			PipelineID:  pipeline.ID,
			TaskID:      task.ID,
			OrgName:     pipeline.OrgName,
			ClusterName: pipeline.ClusterName,
			Stream:      "stdout",
			Tail:        tail,
		}, mode, task.TimeBegin)
		if err != nil {
			return nil, err
		}
		stderrLines, err := fetchTaskLogLines(ctx, common.TaskLogRequestOptions{
			PipelineID:  pipeline.ID,
			TaskID:      task.ID,
			OrgName:     pipeline.OrgName,
			ClusterName: pipeline.ClusterName,
			Stream:      "stderr",
			Tail:        tail,
		}, mode, task.TimeBegin)
		if err != nil {
			return nil, err
		}
		return mergeAndSortTaskLogLines(stdoutLines, stderrLines), nil
	}

	lines, err := fetchTaskLogLines(ctx, common.TaskLogRequestOptions{
		PipelineID:  pipeline.ID,
		TaskID:      task.ID,
		OrgName:     pipeline.OrgName,
		ClusterName: pipeline.ClusterName,
		Stream:      stream,
		Tail:        tail,
	}, mode, task.TimeBegin)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func fetchTaskLogPage(ctx *command.Context, pipeline pipelinepb.PipelineDetailDTO, task common.PipelineTaskRef, stream string, tail int, start int64, count int64) ([]apistructs.DashboardSpotLogLine, error) {
	pageSize := common.DefaultLogPageSize(tail, count)
	if stream == "all" {
		stdoutData, err := fetchTaskLogData(ctx, common.TaskLogRequestOptions{
			PipelineID:  pipeline.ID,
			TaskID:      task.ID,
			OrgName:     pipeline.OrgName,
			ClusterName: pipeline.ClusterName,
			Stream:      "stdout",
			Tail:        tail,
			Start:       start,
			Count:       pageSize,
		})
		if err != nil {
			return nil, err
		}
		stderrData, err := fetchTaskLogData(ctx, common.TaskLogRequestOptions{
			PipelineID:  pipeline.ID,
			TaskID:      task.ID,
			OrgName:     pipeline.OrgName,
			ClusterName: pipeline.ClusterName,
			Stream:      "stderr",
			Tail:        tail,
			Start:       start,
			Count:       pageSize,
		})
		if err != nil {
			return nil, err
		}
		return mergeAndSortTaskLogLines(stdoutData.Lines, stderrData.Lines), nil
	}

	data, err := fetchTaskLogData(ctx, common.TaskLogRequestOptions{
		PipelineID:  pipeline.ID,
		TaskID:      task.ID,
		OrgName:     pipeline.OrgName,
		ClusterName: pipeline.ClusterName,
		Stream:      stream,
		Tail:        tail,
		Start:       start,
		Count:       pageSize,
	})
	if err != nil {
		return nil, err
	}
	return data.Lines, nil
}

func fetchTaskLogLines(ctx *command.Context, opts common.TaskLogRequestOptions, mode taskLogFetchMode, fallbackStart int64) ([]apistructs.DashboardSpotLogLine, error) {
	queryOpts := opts
	for attempt := 0; ; attempt++ {
		data, err := fetchTaskLogData(ctx, queryOpts)
		if err != nil {
			return nil, err
		}
		if len(data.Lines) > 0 {
			return data.Lines, nil
		}
		nextOpts := queryOpts
		if mode.ForwardFallback && canUseForwardWindowFallback(queryOpts, fallbackStart) {
			nextOpts = forwardWindowFallbackOptions(opts, fallbackStart)
		} else if mode.RetryEmpty && attempt < emptyLogFetchAttempts-1 {
			sleepForRetry(emptyLogRetryInterval)
		} else {
			return data.Lines, nil
		}
		queryOpts = nextOpts
	}
}

func canUseForwardWindowFallback(opts common.TaskLogRequestOptions, fallbackStart int64) bool {
	return fallbackStart > 0 && opts.Start == 0 && opts.Count == 0
}

func forwardWindowFallbackOptions(opts common.TaskLogRequestOptions, start int64) common.TaskLogRequestOptions {
	opts.Start = start
	opts.Count = int64(normalizedLogTail(opts.Tail))
	return opts
}

func normalizedLogTail(tail int) int {
	if tail <= 0 {
		return 200
	}
	return tail
}

func writeLogOutput(w io.Writer, pipelineID uint64, tasks []common.PipelineTaskRef, logsByTask map[uint64][]apistructs.DashboardSpotLogLine, opts logOptions) error {
	lines := mergeLogLines(tasks, logsByTask)
	if opts.jsonOutput {
		encoder := json.NewEncoder(w)
		for _, item := range lines {
			record := logOutputRecord{
				PipelineID: pipelineID,
				TaskID:     item.task.ID,
				TaskName:   item.task.Name,
				Stream:     item.line.Stream,
				Timestamp:  item.line.TimeStamp,
				Content:    item.line.Content,
			}
			if err := encoder.Encode(record); err != nil {
				return err
			}
		}
		return nil
	}

	singleTask := len(tasks) == 1
	for _, item := range lines {
		if singleTask {
			if _, err := fmt.Fprintln(w, item.line.Content); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "[%s#%d] %s\n", item.task.Name, item.task.ID, item.line.Content); err != nil {
			return err
		}
	}
	return nil
}

func mergeLogLines(tasks []common.PipelineTaskRef, logsByTask map[uint64][]apistructs.DashboardSpotLogLine) []mergedLogLine {
	taskByID := make(map[uint64]common.PipelineTaskRef, len(tasks))
	for _, task := range tasks {
		taskByID[task.ID] = task
	}

	var merged []mergedLogLine
	for taskID, lines := range logsByTask {
		task := taskByID[taskID]
		for _, line := range lines {
			merged = append(merged, mergedLogLine{task: task, line: line})
		}
	}

	sort.SliceStable(merged, func(i, j int) bool {
		if merged[i].line.TimeStamp == merged[j].line.TimeStamp {
			return merged[i].task.ID < merged[j].task.ID
		}
		return merged[i].line.TimeStamp < merged[j].line.TimeStamp
	})
	return merged
}

func filterUnseenLogLines(taskID uint64, lines []apistructs.DashboardSpotLogLine, seen map[uint64]map[string]struct{}) []apistructs.DashboardSpotLogLine {
	if _, ok := seen[taskID]; !ok {
		seen[taskID] = make(map[string]struct{})
	}

	var filtered []apistructs.DashboardSpotLogLine
	for _, line := range lines {
		key := fmt.Sprintf("%s|%d|%s|%s", line.TimeStamp, line.Offset, line.Stream, line.Content)
		if _, ok := seen[taskID][key]; ok {
			continue
		}
		seen[taskID][key] = struct{}{}
		filtered = append(filtered, line)
	}
	return filtered
}

func defaultSelectedTasks(tasks []common.PipelineTaskRef, mode logSelectionMode, interactive bool) []common.PipelineTaskRef {
	if len(tasks) <= 1 || mode != logSelectionModeAuto || !interactive {
		return tasks
	}

	if task, ok := pickTaskByStatus(tasks, func(status apistructs.PipelineStatus) bool {
		return status.IsRunningStatus()
	}); ok {
		return []common.PipelineTaskRef{task}
	}
	if task, ok := pickTaskByStatus(tasks, func(status apistructs.PipelineStatus) bool {
		return status.IsFailedStatus()
	}); ok {
		return []common.PipelineTaskRef{task}
	}

	latest := tasks[0]
	for _, task := range tasks[1:] {
		if task.TimeBegin > latest.TimeBegin {
			latest = task
		}
	}
	return []common.PipelineTaskRef{latest}
}

func pickTaskByStatus(tasks []common.PipelineTaskRef, predicate func(status apistructs.PipelineStatus) bool) (common.PipelineTaskRef, bool) {
	for _, task := range tasks {
		if predicate(apistructs.PipelineStatus(task.Status)) {
			return task, true
		}
	}
	return common.PipelineTaskRef{}, false
}

func isTTYWriter(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
