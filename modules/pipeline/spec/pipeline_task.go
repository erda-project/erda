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

package spec

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type PipelineTask struct {
	ID         uint64 `json:"id" xorm:"pk autoincr"`
	PipelineID uint64 `json:"pipelineID"`
	StageID    uint64 `json:"stageID"`

	Name         string                        `json:"name"`
	OpType       PipelineTaskOpType            `json:"opType"`         // Deprecated: get, put, task
	Type         string                        `json:"type,omitempty"` // git, buildpack, release, dice ... 当 OpType 为自定义任务时为空
	ExecutorKind PipelineTaskExecutorKind      `json:"executorKind"`   // scheduler, memory
	Status       apistructs.PipelineStatus     `json:"status"`
	Extra        PipelineTaskExtra             `json:"extra" xorm:"json"`
	Context      PipelineTaskContext           `json:"context" xorm:"json"`
	Result       apistructs.PipelineTaskResult `json:"result" xorm:"json"`

	IsSnippet             bool                                  `json:"isSnippet"`                         // 该节点是否是嵌套流水线节点
	SnippetPipelineID     *uint64                               `json:"snippetPipelineID"`                 // 嵌套的流水线 id
	SnippetPipelineDetail *apistructs.PipelineTaskSnippetDetail `json:"snippetPipelineDetail" xorm:"json"` // 嵌套的流水线详情

	CostTimeSec  int64     `json:"costTimeSec"`                // -1 表示暂无耗时信息, 0 表示确实是0s结束
	QueueTimeSec int64     `json:"queueTimeSec"`               // 等待调度的耗时, -1 暂无耗时信息, 0 表示确实是0s结束 TODO 赋值
	TimeBegin    time.Time `json:"timeBegin"`                  // 执行开始时间
	TimeEnd      time.Time `json:"timeEnd"`                    // 执行结束时间
	TimeCreated  time.Time `json:"timeCreated" xorm:"created"` // 记录创建时间
	TimeUpdated  time.Time `json:"timeUpdated" xorm:"updated"` // 记录更新时间
}

func (pt *PipelineTask) NodeName() string {
	return pt.Name
}

func (pt *PipelineTask) PrevNodeNames() []string {
	return pt.Extra.RunAfter
}

func (*PipelineTask) TableName() string {
	return "pipeline_tasks"
}

type PipelineTaskExtra struct {
	Namespace    string                   `json:"namespace,omitempty"`
	ExecutorName PipelineTaskExecutorName `json:"executorName,omitempty"`
	ClusterName  string                   `json:"clusterName,omitempty"`
	AllowFailure bool                     `json:"allowFailure,omitempty"`
	Pause        bool                     `json:"pause,omitempty"`
	Timeout      time.Duration            `json:"timeout,omitempty"`
	PrivateEnvs  map[string]string        `json:"envs,omitempty"`       // PrivateEnvs 由 agent 注入 run 运行时，run 可见，容器内不可见
	PublicEnvs   map[string]string        `json:"publicEnvs,omitempty"` // PublicEnvs 注入容器，run 可见，容器内亦可见
	Labels       map[string]string        `json:"labels,omitempty"`
	Image        string                   `json:"image,omitempty"`
	Cmd          string                   `json:"cmd,omitempty"`
	CmdArgs      []string                 `json:"cmdArgs,omitempty"`
	Binds        []apistructs.Bind        `json:"binds,omitempty"`
	// Volumes 创建 task 时的 volumes 快照
	// 若一开始 volume 无 volumeID，启动 task 后返回的 volumeID 不会在这里更新，只会更新到 task.Context.OutStorages 里
	Volumes         []apistructs.MetadataField `json:"volumes,omitempty"` //
	PreFetcher      *apistructs.PreFetcher     `json:"preFetcher,omitempty"`
	RuntimeResource RuntimeResource            `json:"runtimeResource,omitempty"`
	UUID            string                     `json:"uuid"` // 用于查询日志等，pipeline 开始执行时才会赋值 // 对接多个 executor，不一定每个 executor 都能自定义 UUID，所以这个 uuid 实际上是目标系统的 uuid
	TimeBeginQueue  time.Time                  `json:"timeBeginQueue"`
	TimeEndQueue    time.Time                  `json:"timeEndQueue"`
	StageOrder      int                        `json:"stageOrder"` // 0,1,2,...

	// RunAfter indicates the tasks this task depends.
	RunAfter []string `json:"runAfter"`

	FlinkSparkConf FlinkSparkConf `json:"flinkSparkConf,omitempty"`

	Action pipelineyml.Action `json:"action,omitempty"`

	OpenapiOAuth2TokenPayload apistructs.OpenapiOAuth2TokenPayload `json:"openapiOAuth2TokenPayload"`

	LoopOptions *apistructs.PipelineTaskLoopOptions `json:"loopOptions,omitempty"` // 开始执行后保证不为空

	AppliedResources apistructs.PipelineAppliedResources `json:"appliedResources,omitempty"`

	EncryptSecretKeys []string `json:"encryptSecretKeys"` // the encrypt envs' key list
}

type FlinkSparkConf struct {
	// 该部分在 action 的 source 里声明
	Depend    string   `json:"depends,omitempty"`
	MainClass string   `json:"main_class,omitempty"`
	MainArgs  []string `json:"main_args,omitempty"`

	// flink/spark action 运行需要一个 jar resource（flink 为 jarID，spark 为 jarURL）
	// 该部分在运行期动态赋值
	JarResource string `json:"jarResource,omitempty"`
}

type PipelineTaskContext struct {
	InStorages  apistructs.Metadata `json:"inStorages,omitempty"`
	OutStorages apistructs.Metadata `json:"outStorages,omitempty"`

	CmsDiceFiles apistructs.Metadata `json:"cmsDiceFiles,omitempty"`
}

func (c *PipelineTaskContext) Dedup() {
	c.InStorages = c.InStorages.DedupByName()
	c.OutStorages = c.OutStorages.DedupByName()
}

// GenerateOperation
type PipelineTaskOpType string

var (
	PipelineTaskOpTypeGet  PipelineTaskOpType = "get"
	PipelineTaskOpTypePut  PipelineTaskOpType = "put"
	PipelineTaskOpTypeTask PipelineTaskOpType = "task"
)

type PipelineTaskExecutorKind string

var (
	PipelineTaskExecutorKindScheduler PipelineTaskExecutorKind = "SCHEDULER"
	PipelineTaskExecutorKindMemory    PipelineTaskExecutorKind = "MEMORY"
	PipelineTaskExecutorKindAPITest   PipelineTaskExecutorKind = "APITEST"
	PipelineTaskExecutorKindList                               = []PipelineTaskExecutorKind{PipelineTaskExecutorKindScheduler, PipelineTaskExecutorKindMemory, PipelineTaskExecutorKindAPITest}
)

func (that PipelineTaskExecutorKind) Check() bool {
	for _, kind := range PipelineTaskExecutorKindList {
		if string(kind) == string(that) {
			return true
		}
	}
	return false
}

type PipelineTaskExecutorName string

func (that PipelineTaskExecutorName) String() string {
	return string(that)
}

var (
	PipelineTaskExecutorNameEmpty            PipelineTaskExecutorName = ""
	PipelineTaskExecutorNameSchedulerDefault PipelineTaskExecutorName = "scheduler"
	PipelineTaskExecutorNameAPITestDefault   PipelineTaskExecutorName = "api-test"
	PipelineTaskExecutorNameList                                      = []PipelineTaskExecutorName{PipelineTaskExecutorNameEmpty, PipelineTaskExecutorNameSchedulerDefault, PipelineTaskExecutorNameAPITestDefault}
)

func (that PipelineTaskExecutorName) Check() bool {
	for _, name := range PipelineTaskExecutorNameList {
		if string(name) == string(that) {
			return true
		}
	}
	return false
}

type RuntimeResource struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}

type Volume struct {
	HostPath      string `json:"hostPath"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly"`
}

func (pt *PipelineTask) Convert2DTO() *apistructs.PipelineTaskDTO {
	if pt == nil {
		return nil
	}
	pt.Result.ConvertErrors()
	task := apistructs.PipelineTaskDTO{
		ID:         pt.ID,
		PipelineID: pt.PipelineID,
		StageID:    pt.StageID,
		Name:       pt.Name,
		OpType:     string(pt.OpType),
		Type:       string(pt.Type),
		Status:     pt.Status,
		Extra: apistructs.PipelineTaskExtra{
			UUID:         pt.Extra.UUID,
			AllowFailure: pt.Extra.AllowFailure,
		},
		Labels:       pt.Extra.Action.Labels,
		Result:       pt.Result,
		CostTimeSec:  pt.CostTimeSec,
		QueueTimeSec: pt.QueueTimeSec,
		TimeBegin:    pt.TimeBegin,
		TimeEnd:      pt.TimeEnd,
		TimeCreated:  pt.TimeCreated,
		TimeUpdated:  pt.TimeUpdated,

		IsSnippet:             pt.IsSnippet,
		SnippetPipelineID:     pt.SnippetPipelineID,
		SnippetPipelineDetail: pt.SnippetPipelineDetail,
	}
	// handle metadata
	for _, field := range task.Result.Metadata {
		field.Level = field.GetLevel()
	}

	if task.Status.IsSuccessStatus() {
		task.Result.Errors = nil
		notErrorMeta, _ := task.Result.Metadata.FilterNoErrorLevel()
		task.Result.Metadata = notErrorMeta
	}

	if task.Type == "manual-review" {
		task.Status = task.Status.ChangeStateForManualReview()
	}

	return &task
}

func (pt *PipelineTask) RuntimeID() string {
	for _, meta := range pt.Result.Metadata {
		if meta.Type == apistructs.ActionCallbackTypeLink &&
			meta.Name == apistructs.ActionCallbackRuntimeID {
			return meta.Value
		}
	}
	return ""
}

func (pt *PipelineTask) ReleaseID() string {
	for _, meta := range pt.Result.Metadata {
		if meta.Type == apistructs.ActionCallbackTypeLink &&
			meta.Name == apistructs.ActionCallbackReleaseID {
			return meta.Value
		}
	}
	return ""
}

func GenDefaultTaskResource() RuntimeResource {
	return RuntimeResource{
		CPU:    conf.TaskDefaultCPU(),
		Memory: conf.TaskDefaultMEM(),
		Disk:   0,
	}
}
