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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindFirstPlaceholder(t *testing.T) {
	diceKey := "diceKey: ((dice.projectID)):((dice.env)):((dice.branch)):((dice.operatorID))"
	placeholder, find := findFirstPlaceholder(diceKey)
	require.True(t, find)
	require.Equal(t, "((dice.projectID))", placeholder)

	normal := "diceKey: 666:test:master:92"
	placeholder, find = findFirstPlaceholder(normal)
	require.False(t, find)

	odd := "diceKey: ))((1)):test:master:92"
	placeholder, find = findFirstPlaceholder(odd)
	require.True(t, find)
	require.Equal(t, "((1))", placeholder)
}

func TestPipelineYml_Evaluate(t *testing.T) {
	ymlContent :=
		`
version: 1.0
# ((noneed))
stages: # stage begin...
- name: deploy
  tasks:
    - put: ((put-resource-name))
      params:
        diceKey: ((dice.projectID)):((dice.env)):((dice.branch)):((dice.operatorID))

- name: bp-backend
  type: buildpack
  source:
    context: repo/services
    bp_repo: http://git.terminus.io/buildpacks/dice-bpack-termjava.git
    bp_ver: feature/new-bp
    modules:
    - name: galaxy-admin
      image: { name: ))()()((buildpack.image.galaxy-admin)))()()) }
    - name: galaxy-web
      image: { name: ((buildpack.image.galaxy-web)) }
    - name: galaxy-trade
      image: { name: ((buildpack.image.galaxy-trade)) }
    - name: galaxy-item
      image: { name: ((buildpack.image.galaxy-item)) }
    - name: galaxy-user
      image: { name: ((buildpack.image.galaxy-user)) }
`
	y := New([]byte(ymlContent))
	fmt.Println(string(y.byteData))
	variables := map[string]string{
		"((put-resource-name))":            "deploy",
		"((dice.projectID))":               "666",
		"((dice.env))":                     "test",
		"((dice.operatorID))":              "92",
		"((dice.branch))":                  "feature/pipeline",
		"((buildpack.image.galaxy-admin))": "localhost:5000/galaxy/galaxy-admin:v0.1",
		"((buildpack.image.galaxy-web))":   "localhost:5000/galaxy/galaxy-web:v0.1",
		"((buildpack.image.galaxy-trade))": "localhost:5000/galaxy/galaxy-trade:v0.1",
		"((buildpack.image.galaxy-item))":  "localhost:5000/galaxy/galaxy-item:v0.1",
		"((buildpack.image.galaxy-user))":  "localhost:5000/galaxy/galaxy-user:v0.1",
	}
	err := y.evaluate(Map2MetadataFields(variables))
	require.NoError(t, err)
	fmt.Println(string(y.byteData))
}

func TestRemoveComment(t *testing.T) {
	s := "((dice.id#))#"
	nocomment, comment := removeComment(s)
	require.Equal(t, "((dice.id", nocomment)
	require.Equal(t, "#))#", comment)
}

func TestRenderPlaceholders(t *testing.T) {
	s := "((dice.id)) . ((a)) #comment# ((noneed))"
	output, err := RenderPlaceholders(s, Map2MetadataFields(map[string]string{"((dice.id))": "id", "((a))": "a"}))
	require.NoError(t, err)
	fmt.Println(output)
}
