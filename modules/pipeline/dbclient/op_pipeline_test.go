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

package dbclient

//import (
//	"fmt"
//	"os"
//	"testing"
//	"time"
//
//	"github.com/davecgh/go-spew/spew"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//
//	"github.com/erda-project/erda/apistructs"
//	"github.com/erda-project/erda/modules/pipeline/spec"
//)
//
//var (
//	client *Client
//	err    error
//)
//
//func init() {
//	_ = os.Setenv("MYSQL_DATABASE", "dice")
//	_ = os.Setenv("MYSQL_SHOW_SQL", "true")
//	client, err = New()
//	if err != nil {
//		panic(err)
//	}
//}
//
//func TestClient_PageListPipelines(t *testing.T) {
//	pipelines, _, total, currentPageSize, err := client.PageListPipelines(apistructs.PipelinePageListRequest{
//		AppID:    456,
//		Sources:  []apistructs.PipelineSource{"dice"},
//		Branches: []string{"develop"},
//		YmlNames: []string{"pipeline.yml", "path1/path2/demo.workflow"},
//		PageSize: 5,
//		PageNum:  1,
//	})
//	require.NoError(t, err)
//	fmt.Println("len pipelines =", len(pipelines))
//	fmt.Println("total = ", total)
//	fmt.Println("currentPageSize = ", currentPageSize)
//}
//
//func TestClient_PageListPipelinesFDP(t *testing.T) {
//	pipelines, _, total, currentPageSize, err := client.PageListPipelines(apistructs.PipelinePageListRequest{
//		Sources: []apistructs.PipelineSource{apistructs.PipelineSourceCDPDev},
//		Statuses: []string{
//			apistructs.PipelineStatusRunning.String(),
//			apistructs.PipelineStatusFailed.String(),
//			apistructs.PipelineStatusStopByUser.String(),
//			apistructs.PipelineStatusSuccess.String(),
//		},
//		ClusterNames:   []string{"scjumax-prod"},
//		StartTimeBegin: time.Date(2020, 11, 24, 23, 0, 0, 0, time.Local),
//		LargePageSize:  true,
//		CountOnly:      false,
//		SelectCols:     nil,
//		AscCols:        nil,
//		DescCols:       []string{"id"},
//	})
//	require.NoError(t, err)
//	fmt.Println("len pipelines =", len(pipelines))
//	fmt.Println("total = ", total)
//	fmt.Println("currentPageSize = ", currentPageSize)
//}
//
//func TestClient_PipelineStatisticFDP(t *testing.T) {
//	data, err := client.PipelineStatistic("cdp-dev", "scjumax-prod")
//	assert.NoError(t, err)
//	spew.Dump(data)
//}
//
//func TestClient_PageListPipelinesWithLabel(t *testing.T) {
//	pipelines, _, total, currentPageSize, err := client.PageListPipelines(apistructs.PipelinePageListRequest{
//		Sources:        []apistructs.PipelineSource{apistructs.PipelineSourceDefault, apistructs.PipelineSourceDice},
//		AllSources:     true,
//		AnyMatchLabels: map[string][]string{"k1": {"v1"}},
//	})
//	require.NoError(t, err)
//	fmt.Println("len pipelines =", len(pipelines))
//	fmt.Println("total = ", total)
//	fmt.Println("currentPageSize = ", currentPageSize)
//}
//
//func TestClient_ListInvokedCombos(t *testing.T) {
//	combos, err := client.ListAppInvokedCombos(1,
//		spec.PipelineCombosReq{
//			Branches: []string{"master", "not-exist"},
//			// YmlNames: []string{"pipeline.yml"},
//			Sources: []string{"bigdata", "dice"},
//		})
//	require.NoError(t, err)
//	for _, combo := range combos {
//		fmt.Printf("%v\n", combo)
//	}
//}
//
//func TestClient_GetPipeline(t *testing.T) {
//	_, found, err := client.GetPipelineWithExistInfo(84)
//	assert.NoError(t, err)
//	assert.False(t, found)
//}
//
//func TestClient_DeletePipeline(t *testing.T) {
//	p := spec.Pipeline{}
//
//	// no tx
//	fmt.Println("no tx begin")
//
//	err := client.CreatePipeline(&p)
//	fmt.Println(p.ID)
//	require.NoError(t, err)
//
//	err = client.DeletePipeline(p.ID)
//	require.NoError(t, err)
//
//	_, err = client.GetPipeline(p.ID)
//	require.Error(t, err)
//	fmt.Println("no tx end")
//
//	// tx commit
//	p = spec.Pipeline{}
//	fmt.Println("tx begin")
//	txSession := client.Engine.NewSession()
//	err = txSession.Begin()
//	require.NoError(t, err)
//
//	err = client.CreatePipeline(&p, WithTxSession(txSession))
//	fmt.Println(p.ID)
//	require.NoError(t, err)
//
//	// not commit yet
//	_, err = client.GetPipeline(p.ID) // not commit, so not exist
//	require.Error(t, err)
//
//	// commit
//	err = txSession.Commit()
//	require.NoError(t, err)
//
//	_p, err := client.GetPipeline(p.ID) // not commit, so not exist
//	require.NoError(t, err)
//	require.Equal(t, p.ID, _p.ID)
//
//	fmt.Println("tx end")
//}
//
//func TestClient_NewSession(t *testing.T) {
//	session := client.NewSession()
//	session.Close()
//	session.Close()
//}
