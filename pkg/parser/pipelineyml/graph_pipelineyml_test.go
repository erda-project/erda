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

package pipelineyml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertGraphPipelineYml(t *testing.T) {
	fb, err := ioutil.ReadFile("./samples/graph_pipelineyml.yaml")
	assert.NoError(t, err)
	b, err := ConvertGraphPipelineYmlContent(fb)
	assert.NoError(t, err)
	fmt.Println(string(b))
}

func TestConvertToGraphPipelineYml(t *testing.T) {
	fb, err := ioutil.ReadFile("./samples/pipeline_cicd.yml")
	assert.NoError(t, err)
	graph, err := ConvertToGraphPipelineYml(fb)
	assert.NoError(t, err)
	b, err := json.MarshalIndent(graph, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(b))
}
