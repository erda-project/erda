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

func TestUpgradeYmlFromV1(t *testing.T) {
	b, err := ioutil.ReadFile("./pipelineymlv1/samples/pipeline.yml")
	assert.NoError(t, err)
	newYmlByte, err := UpgradeYmlFromV1(b)
	assert.NoError(t, err)
	fmt.Println(string(newYmlByte))

}

func TestUpgradeYmlFromV1_PampasBlog(t *testing.T) {
	b, err := ioutil.ReadFile("./pipelineymlv1/samples/pipeline-pampas-blog.yml")
	assert.NoError(t, err)
	newYmlByte, err := UpgradeYmlFromV1(b)
	assert.NoError(t, err)
	fmt.Println(string(newYmlByte))
}
