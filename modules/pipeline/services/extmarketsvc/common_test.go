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

package extmarketsvc

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"

	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func TestMakeActionTypeVersion(t *testing.T) {
	item := MakeActionTypeVersion(&pipelineyml.Action{Type: "git", Version: "1.0"})
	assert.Equal(t, item, "git@1.0")

	item = MakeActionTypeVersion(&pipelineyml.Action{Type: "git"})
	assert.Equal(t, item, "git")
}
