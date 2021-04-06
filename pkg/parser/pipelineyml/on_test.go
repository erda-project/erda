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
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOn(t *testing.T) {
	b, err := ioutil.ReadFile("./samples/on.yml")
	assert.NoError(t, err)
	newYml, err := ConvertToGraphPipelineYml(b)
	assert.NoError(t, err)
	assert.Contains(t, newYml.YmlContent, "push")
	//n := &apistructs.PipelineYml{}
	//assert.EqualValues(t, newYml.On, n.On)
	fmt.Println(newYml.YmlContent)
}
