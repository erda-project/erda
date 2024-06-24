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

package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/action_info"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cache"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgereporter"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/mock"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type CronServiceServerTestImpl struct {
	create *cronpb.CronCreateResponse
}

func (c CronServiceServerTestImpl) CronCreate(ctx context.Context, request *cronpb.CronCreateRequest) (*cronpb.CronCreateResponse, error) {
	return c.create, nil
}

func (c CronServiceServerTestImpl) CronPaging(ctx context.Context, request *cronpb.CronPagingRequest) (*cronpb.CronPagingResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronStart(ctx context.Context, request *cronpb.CronStartRequest) (*cronpb.CronStartResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronStop(ctx context.Context, request *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronDelete(ctx context.Context, request *cronpb.CronDeleteRequest) (*cronpb.CronDeleteResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronGet(ctx context.Context, request *cronpb.CronGetRequest) (*cronpb.CronGetResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronUpdate(ctx context.Context, request *cronpb.CronUpdateRequest) (*cronpb.CronUpdateResponse, error) {
	panic("implement me")
}

func TestPipelineSvc_UpdatePipelineCron(t *testing.T) {
	type args struct {
		p                      *spec.Pipeline
		cronStartFrom          *timestamppb.Timestamp
		configManageNamespaces []string
		cronCompensator        *pipelineyml.CronCompensator
	}
	tests := []struct {
		name    string
		args    args
		isEdge  bool
		wantErr bool
	}{
		{
			name: "test id > 0",
			args: args{
				p: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						PipelineSource:  "test",
						PipelineYmlName: "test",
						ClusterName:     "test",
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CronExpr: "test",
						},
						PipelineYml: "test",
						Snapshot: spec.Snapshot{
							Envs: map[string]string{
								"test": "test",
							},
						},
					},
					Labels: map[string]string{
						"test": "test",
					},
				},
				cronStartFrom:          nil,
				configManageNamespaces: nil,
				cronCompensator:        nil,
			},
			isEdge:  false,
			wantErr: false,
		},
		{
			name: "test edge cron",
			args: args{
				p: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						PipelineSource:  "test",
						PipelineYmlName: "test",
						ClusterName:     "test",
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CronExpr: "test",
						},
						PipelineYml: "test",
						Snapshot: spec.Snapshot{
							Envs: map[string]string{
								"test": "test",
							},
						},
					},
					Labels: map[string]string{
						"test": "test",
					},
				},
				cronStartFrom:          nil,
				configManageNamespaces: nil,
				cronCompensator:        nil,
			},
			isEdge:  true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &pipelineService{}
			impl := CronServiceServerTestImpl{
				create: &cronpb.CronCreateResponse{
					Data: &pb.Cron{
						ID: 1,
					},
				},
			}
			s.cronSvc = impl
			if tt.isEdge {
				s.edgeRegister = &edgepipeline_register.MockEdgeRegister{}
				s.edgeReporter = &edgereporter.MockEdgeReporter{}
			}

			if err := s.UpdatePipelineCron(tt.args.p, tt.args.cronStartFrom, tt.args.configManageNamespaces, tt.args.cronCompensator); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePipelineCron() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

var (
	successCode         = 200
	limitCostTime int64 = 500 // ms
)

type benchmarkPipelineCreateRes struct {
	Name       string `json:"name"`
	CostTime   int64  `json:"costTime"` // ms
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
	err        error  `json:"err"`
}

type benchmarkQueueCreateRes struct {
	Name       string `json:"name"`
	QueueID    int    `json:"queueID"`
	StatusCode int    `json:"statusCode"`
	err        error  `json:"err"`
	Data       struct {
		ID int `json:"id"`
	}
}

func createV2(url string) (result benchmarkPipelineCreateRes) {
	if len(url) == 0 {
		return
	}
	method := "POST"

	payload := strings.NewReader(`{
    "appID": 10,
    "pipelineYml": "version: \"1.1\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          description: 代码仓库克隆\n          version: \"1.0\"\n          params:\n            branch: ((gittar.branch))\n            depth: 1\n            password: ((gittar.password))\n            uri: ((gittar.repo))\n            username: ((gittar.username))\n          timeout: 3600\n  - stage:\n      - golang:\n          alias: go-demo\n          description: golang action\n          params:\n            command: go build -o web-server main.go\n            context: ${git-checkout}\n            service: web-server\n            target: web-server\n  - stage:\n      - release:\n          alias: release\n          description: 用于打包完成时，向dicehub 提交完整可部署的dice.yml。用户若没在pipeline.yml里定义该action，CI会自动在pipeline.yml里插入该action\n          params:\n            dice_yml: ${git-checkout}/dice.yml\n            image:\n              go-demo: ${go-demo:OUTPUT:image}\n  - stage:\n      - dice:\n          alias: dice\n          description: 用于 Erda 平台部署应用服务\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n",
    "pipelineSource": "dice",
    "pipelineYmlName": "pipeline.yml",
    "clusterName": "dev",
    "autoRunAtOnce": false,
    "autoStartCron": false,
    "labels": {
        "branch": "develop",
        "diceWorkspace": "TEST",
        "has-report-basic": "true"
    },
    "normalLabels": {
        "appName": "go-demo",
        "projectName": "go-demo",
        "orgName": "erda"
    },
    "configManageNamespaces": []
}`)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		result.err = err
		return
	}
	req.Header.Add("Internal-Client", "bundle")
	req.Header.Add("Content-Type", "application/json")

	timeStart := time.Now()
	res, err := client.Do(req)
	if err != nil {
		result.err = err
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		result.err = err
		return
	}
	timeEnd := time.Now()
	result.CostTime = timeEnd.Sub(timeStart).Milliseconds()
	result.StatusCode = res.StatusCode
	result.Body = string(body)
	result.Name = "createV2"
	return
}

func newCreateV2Work(workerID int) (count int) {
	resChan := make(chan benchmarkPipelineCreateRes, 0)
	execChan := make(chan struct{}, 0)
	defer close(resChan)
	defer close(execChan)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case <-execChan:
				data := createV2("")
				resChan <- data
			case <-ctx.Done():
				return
			}
		}
	}()
	execChan <- struct{}{}
	for {
		select {
		case data := <-resChan:
			count++
			logrus.Infof("worker: %d execute pipeline, statusCode: %d, costTime: %dms", workerID, data.StatusCode, data.CostTime)
			if data.err != nil {
				logrus.Errorf("%s: %s", data.Name, data.err.Error())
				return
			}
			if data.StatusCode != successCode {
				logrus.Errorf("status code is not %d, but %d", successCode, data.StatusCode)
				return
			}
			if count >= 1 {
				return
			}
			execChan <- struct{}{}
		}
	}
	return
}

//func Test_benchmarkCreateV2(t *testing.T) {
//	wait := &sync.WaitGroup{}
//	// concurrency of create pipeline
//	for i := 0; i < 1; i++ {
//		wait.Add(1)
//		go func(idx int) {
//			defer wait.Done()
//			count := newCreateV2Work(idx)
//			logrus.Infof("worker %d created %d pipelines", idx, count)
//		}(i)
//	}
//	wait.Wait()
//}

func createFDPPipelineAutoRun(idx int, url string) (result benchmarkPipelineCreateRes) {
	if len(url) == 0 {
		return
	}
	method := "POST"

	payload := `{
    "pipelineYml": "version: \"1.1\"\nname: \"\"\nstages:\n  - stage:\n      - custom-script:\n          alias: custom-script\n          version: \"1.0\"\n          commands:\n            - sleep 600\n          resources:\n            cpu: 0.5\n            mem: 1024\nlifecycle:\n  - hook: before-run-check\n    client: FDP\n    labels:\n      clusterName: dy-prod\n      dependWorkflowIds:\n        - dy-prod-10\n      workflowId: dy-prod-143",
    "clusterName": "csi-dev",
	"autoRunAtOnce": true,
    "pipelineYmlName": "enqueue-custom-%d",
    "pipelineSource": "cdp-dev",
    "labels": {
        "appID": "3",
        "orgID": "1",
        "projectID": "3"
    },
    "normalLabels": {
        "appName": "benchmark",
        "orgName": "benchmark",
        "projectName": "benchmark"
    }
}`
	payload = fmt.Sprintf(payload, idx)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/v2/pipelines", url), strings.NewReader(payload))

	if err != nil {
		result.err = err
		return
	}
	req.Header.Add("Internal-Client", "bundle")
	req.Header.Add("Content-Type", "application/json")

	timeStart := time.Now()
	res, err := client.Do(req)
	if err != nil {
		result.err = err
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		result.err = err
		return
	}
	timeEnd := time.Now()
	result.CostTime = timeEnd.Sub(timeStart).Milliseconds()
	result.StatusCode = res.StatusCode
	result.Body = string(body)
	result.Name = "createV2"
	return
}

func createQueue(idx int, url string) (result benchmarkQueueCreateRes) {
	method := "POST"

	payload := `{
    "name": "enqueue-test-%d",
    "pipelineSource": "cdp-dev",
    "clusterName":"csi-dev",
    "concurrency": 1000,
    "maxCPU": 3000,
    "maxMemoryMB": 40960000
}`
	payload = fmt.Sprintf(payload, idx)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(method, fmt.Sprintf("%s/api/pipeline-queues", url), strings.NewReader(payload))

	if err != nil {
		result.err = err
		return
	}
	req.Header.Add("Internal-Client", "bundle")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		result.err = err
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		result.err = err
		return
	}
	if err := json.Unmarshal(body, &result); err != nil {
		result.err = err
		return
	}
	result.StatusCode = res.StatusCode
	result.Name = "createQueue"
	result.QueueID = result.Data.ID
	return
}

//func Test_benchmarkCreateAndRunPipelineInQueue(t *testing.T) {
//	for i := 0; i < 1; i++ {
//		res := createFDPPipelineAutoRun(i, "http://localhost:3081")
//		if res.err != nil {
//			t.Errorf("failed to create queue pipeline, err: %v", res.err)
//		}
//		t.Logf("create queue pipeline %d, statusCode: %d, costTime: %dms", i, res.StatusCode, res.CostTime)
//		time.Sleep(time.Second)
//	}
//}

type mockLogger struct {
	mock.MockLogger
}

func (m *mockLogger) Debugf(template string, args ...interface{}) {}

type mockCache struct{}

func (m *mockCache) GetOrSetOrgName(orgID uint64) string {
	if orgID == 1 {
		return "erda"
	}
	return ""
}
func (m *mockCache) GetOrSetPipelineRerunSuccessTasksFromContext(pipelineID uint64) (successTasks map[string]*spec.PipelineTask, err error) {
	return nil, nil
}
func (m *mockCache) GetOrSetStagesFromContext(pipelineID uint64) (stages []spec.PipelineStage, err error) {
	return nil, nil
}
func (m *mockCache) GetOrSetPipelineYmlFromContext(pipelineID uint64) (yml *pipelineyml.PipelineYml, err error) {
	return nil, nil
}
func (m *mockCache) GetOrSetPassedDataWhenCreateFromContext(pipelineYml *pipelineyml.PipelineYml, pipeline *spec.Pipeline) (passedDataWhenCreate *action_info.PassedDataWhenCreate, err error) {
	return nil, nil
}
func (m *mockCache) ClearReconcilerPipelineContextCaches(pipelineID uint64)                     {}
func (m *mockCache) SetPipelineSecretByPipelineID(pipelineID uint64, secret *cache.SecretCache) {}
func (m *mockCache) GetPipelineSecretByPipelineID(pipelineID uint64) (secret *cache.SecretCache) {
	return nil
}
func (m *mockCache) ClearPipelineSecretByPipelineID(pipelineID uint64) {}

func Test_tryGetOrgName(t *testing.T) {
	testCases := []struct {
		name string
		p    *spec.Pipeline
		want string
	}{
		{
			name: "valid orgID",
			p: &spec.Pipeline{
				PipelineExtra: spec.PipelineExtra{
					NormalLabels: map[string]string{
						apistructs.LabelOrgID: "1",
					},
				},
			},
			want: "erda",
		},
		{
			name: "invalid orgID",
			p: &spec.Pipeline{
				PipelineExtra: spec.PipelineExtra{
					NormalLabels: map[string]string{
						apistructs.LabelOrgID: "invalid",
					},
				},
			},
			want: "",
		},
	}
	s := &pipelineService{
		p: &provider{
			Log: &mockLogger{},
		},
		cache: &mockCache{},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orgName := s.tryGetOrgName(tc.p)
			assert.Equal(t, tc.want, orgName)
		})
	}
}
