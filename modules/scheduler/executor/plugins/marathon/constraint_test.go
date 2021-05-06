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

package marathon

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstraints(t *testing.T) {
	cons := NewConstraints()

	r1 := cons.NewLikeRule("TATATAT")
	r1.OR(AND("xxx", "yyy"), AND("platform"))

	reg := regexp.MustCompile(r1.generate())

	assert.Zero(t, len(reg.FindStringIndex("xxx")))
	assert.NotZero(t, len(reg.FindStringIndex("xxx,yyy")))
	assert.NotZero(t, len(reg.FindStringIndex("xxx,platform")))

	r2 := cons.NewLikeRule("dice_tags")
	r2.OR(AND("non"), AND("org-terminus", "workspace-dev", "workspace-test"))
	reg2 := regexp.MustCompile(r2.generate())

	assert.NotZero(t,
		len(reg2.FindStringIndex("any,org-terminus,workspace-dev,workspace-test,workspace-staging,workspace-produn")))
}
