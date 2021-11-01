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

package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCICDPipelineListRequest_GetPageNo(t *testing.T) {
	// none => 0
	r1 := CICDPipelineListRequest{}
	assert.Equal(t, 0, r1.EnsurePageNo())
	assert.Equal(t, r1.PageNo, r1.PageNum)

	// use pageNo
	r2 := CICDPipelineListRequest{PageNo: 1, PageNum: 2}
	assert.Equal(t, 1, r2.EnsurePageNo())
	assert.Equal(t, r2.PageNo, r2.PageNum)

	// use pageNum
	r3 := CICDPipelineListRequest{PageNo: 0, PageNum: 2}
	assert.Equal(t, 2, r3.EnsurePageNo())
	assert.Equal(t, r3.PageNo, r3.PageNum)
}
