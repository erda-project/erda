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

package bundle

//import (
//	"fmt"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/davecgh/go-spew/spew"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestBundle_CreatePipeline(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "localhost:3081")
//	b := New(WithPipeline())
//	_, err := b.CreatePipeline(&apistructs.PipelineCreateRequest{
//		AppID:             1,
//		Branch:            "release/3.3",
//		Source:            "dice",
//		PipelineYmlSource: "gittar",
//		PipelineYmlName:   "pipeline.yml",
//		UserID:            "2",
//		AutoRun:           false,
//	})
//	assert.NoError(t, err)
//}
//
//func TestBundle_Pipeline(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "pipeline.default.svc.cluster.local:3081")
//	b := New(WithPipeline())
//
//	var (
//		p      *apistructs.PipelineDTO
//		detail *apistructs.PipelineDetailDTO
//		err    error
//	)
//
//	// create
//	p, err = b.CreatePipeline(&apistructs.PipelineCreateRequestV2{
//		PipelineYml: `
//version: 1.1
//stages:
//- stage:
//  - custom-script:
//      commands:
//      - echo hello
//`,
//		ClusterName:            "terminus-test",
//		PipelineYmlName:        "bundle_test.yml",
//		PipelineSource:         apistructs.PipelineSourceDefault,
//		Labels:                 nil,
//		NormalLabels:           nil,
//		Envs:                   nil,
//		ConfigManageNamespaces: nil,
//		AutoRun:                false,
//		AutoRunAtOnce:          false,
//		AutoStartCron:          false,
//		CronStartFrom:          nil,
//	})
//	assert.NoError(t, err)
//	fmt.Printf("create pipeline id: %d\n", p.ID)
//
//	//createdPipelineID := p.ID
//
//	// get
//	detail, err = b.GetPipeline(p.ID)
//	assert.NoError(t, err)
//	assert.True(t, p.ID == detail.ID)
//
//	// run
//	//err = b.RunPipeline(p.ID)
//	//assert.NoError(t, err)
//
//	time.Sleep(time.Second * 10)
//
//	// cancel
//	//err = b.CancelPipeline(p.ID)
//	//assert.NoError(t, err)
//
//	// rerun-failed
//	//p, err = b.RerunFailedPipeline(createdPipelineID, false)
//	//assert.NoError(t, err)
//	//fmt.Printf("rerun-failed pipeline id: %d\n", p.ID)
//
//	// rerun
//	//p, err = b.RerunPipeline(createdPipelineID, true)
//	//assert.NoError(t, err)
//	//fmt.Printf("rerun(auto run) pipeline id: %d\n", p.ID)
//
//	// page
//	//pageData, err := b.PageListPipeline(apistructs.PipelinePageListRequest{
//	//	Sources:  "default",
//	//	YmlNames: "bundle_test.yml",
//	//	PageNum:  1,
//	//	PageSize: 10,
//	//})
//	//assert.NoError(t, err)
//	//spew.Dump(pageData)
//}
//
//func TestBundle_PageListPipeline(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "localhost:3081")
//	b := New(WithPipeline())
//	pageData, err := b.PageListPipeline(apistructs.PipelinePageListRequest{
//		Sources:      []apistructs.PipelineSource{apistructs.PipelineSourceDice},
//		ClusterNames: []string{"terminus-xxx"},
//		PageNum:      1,
//		PageSize:     1,
//	})
//	assert.NoError(t, err)
//	for _, p := range pageData.Pipelines {
//		fmt.Println(p.ClusterName)
//	}
//}
//
//func TestBundle_ParsePipelineYmlGraph(t *testing.T) {
//	os.Setenv("PIPELINE_ADDR", "10.125.122.208:3081")
//	b := New(WithPipeline())
//	p, err := b.ParsePipelineYmlGraph(apistructs.PipelineYmlParseGraphRequest{PipelineYmlContent: "version: 1.1"})
//	assert.NoError(t, err)
//	spew.Dump(p)
//}
