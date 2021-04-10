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

package issueFilter

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestComponentFilter_ImportExport(t *testing.T) {
	c := apistructs.Component{
		State: map[string]interface{}{
			"title": "new title",
		},
	}

	f := ComponentFilter{}
	err := f.Import(&c)
	assert.NoError(t, err)
	cc, err := f.SetToProtocolComponent()
	assert.NoError(t, err)
	spew.Dump(cc)
}
