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

package pipelineymlv1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipelineYml_GetTriggerScheduleCron(t *testing.T) {
	y := New([]byte(
		`
version: '1.0'

triggers:
- schedule:
    cron: "* * * * *"
    filters:
    - type: git-branch
      onlys:
      - master
- schedule:
    cron: "*/5 * * * *"
    filters:
    - type: git-branch
      excepts:
      - test
`))
	err := y.Parse(WithBranch("master"))
	require.Error(t, err)

	err = y.Parse(WithBranch("develop"))
	require.NoError(t, err)
}
