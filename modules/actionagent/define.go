// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package actionagent

import (
	"context"
	"os"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/actionagent/filewatch"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

const (
	CONTEXTDIR = "CONTEXTDIR"
	WORKDIR    = "WORKDIR"
	METAFILE   = "METAFILE"
	UPLOADDIR  = "UPLOADDIR"
)

type Agent struct {
	Arg *AgentArg

	EasyUse EasyUse

	// errs collect occurred errors
	Errs []error

	FileWatcher *filewatch.Watcher

	// PushedMetaFileMap store key value already pushed
	PushedMetaFileMap     map[string]string
	LockPushedMetaFileMap sync.RWMutex

	Ctx      context.Context
	Cancel   context.CancelFunc // cancel when logic done
	ExitCode int
}

type AgentArg struct {
	PullBootstrapInfo bool `json:"pullBootstrapInfo"`

	Commands []string                 `json:"commands,omitempty"` // custom action commands -> script -> run
	Context  spec.PipelineTaskContext `json:"context,omitempty"`  // 上下文

	PrivateEnvs map[string]string `json:"privateEnvs,omitempty"`

	PipelineID     uint64 `json:"pipelineID"`
	PipelineTaskID uint64 `json:"pipelineTaskID"`
}

type EasyUse struct {
	ContainerContext  string // 容器内程序运行时上下文目录
	ContainerWd       string // 执行 run 程序时所在目录
	ContainerMetaFile string // metadata 文件

	ContainerUploadDir        string // uploadDir，该目录下的文件在执行结束后会自动上传，并提供下载
	ContainerTempTarUploadDir string // temp tar dir，需要 prepare 时预先创建，用于存放 upload 生成的 tar

	IsEdgeCluster bool // 是否是边缘集群

	RunScript              string // run 文件
	RunProcess             *os.Process
	RunMultiStdoutFilePath string   // multiWriter(os.Stdout) 的文件路径
	RunMultiStdout         *os.File // multiWriter(os.Stdout) 的文件
	RunMultiStderrFilePath string   // multiWriter(os.Stderr) 的文件路径
	RunMultiStderr         *os.File // multiWriter(os.Stderr) 的文件

	OpenAPIAddr       string
	OpenAPIToken      string
	TokenForBootstrap string

	EnablePushLog2Collector bool   // 是否推送日志到 collector
	CollectorAddr           string // collector 地址
	TaskLogID               string // 日志 ID，推送和查询时需要一致

	// Machine stat
	MachineStat apistructs.PipelineTaskMachineStat
}

type RunningEnvironment struct {
	HostIP string
}

func NewAgentArgForPull(pipelineID, pipelineTaskID uint64) *AgentArg {
	var req AgentArg
	req.PullBootstrapInfo = true
	req.PipelineID = pipelineID
	req.PipelineTaskID = pipelineTaskID
	return &req
}

func (agent *Agent) AppendError(err error) {
	if err == nil {
		return
	}
	agent.ExitCode = 1
	agent.Errs = append(agent.Errs, err)
}

func (agent *Agent) MergeErrors() []apistructs.ErrorResponse {
	if len(agent.Errs) == 0 {
		return nil
	}
	agent.ExitCode = 1
	var errs []apistructs.ErrorResponse
	for _, err := range agent.Errs {
		errs = append(errs, apistructs.ErrorResponse{Msg: err.Error()})
	}
	return errs
}
