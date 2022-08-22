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

func Test_ticketTarget(t *testing.T) {
	app := TicketApp
	assert.Equal(t, "application", app.String())

	cluster := TicketCluster
	assert.Equal(t, "cluster", cluster.String())
}

func Test_ticketType(t *testing.T) {
	codeSmell := TicketCodeSmell
	assert.Equal(t, "codeSmell", codeSmell.String())

	bug := TicketBug
	assert.Equal(t, "bug", bug.String())
}
