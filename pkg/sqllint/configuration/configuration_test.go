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

package configuration_test

import (
	"io/ioutil"
	"testing"

	"github.com/erda-project/erda/pkg/sqllint/configuration"
)

func TestFromData(t *testing.T) {
	var filename = "../testdata/config.yaml"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	c, err := configuration.FromData(data)
	if err != nil {
		t.Fatal(err)
	}
	assert(c, t)
}

func TestFromLocal(t *testing.T) {
	var filename = "../testdata/config.yaml"
	c, err := configuration.FromLocal(filename)
	if err != nil {
		t.Fatal(err)
	}
	assert(c, t)
}

func TestConfiguration_ToJsonIndent(t *testing.T) {
	c := new(configuration.Configuration)
	c.AllowedDDLs = new(configuration.AllowedDDLs)
	c.AllowedDMLs = new(configuration.AllowedDMLs)
	_, err := c.ToJsonIndent()
	if err != nil {
		t.Fatal(err)
	}
}

func TestConfiguration_ToYaml(t *testing.T) {
	c := new(configuration.Configuration)
	c.AllowedDDLs = new(configuration.AllowedDDLs)
	c.AllowedDMLs = new(configuration.AllowedDMLs)
	_, err := c.ToYaml()
	if err != nil {
		t.Fatal(err)
	}
}

func assert(c *configuration.Configuration, t *testing.T) {
	if c.AllowedDDLs == nil {
		t.Fatal("failed to FromLocal, c.AllowedDDL should not be nil")
	}
	if c.AllowedDMLs == nil {
		t.Fatal("failed to FromLocal, c.AllowedDML should not be nil")
	}
	if !c.BooleanFieldLinter {
		t.Fatal("failed to FromLocal")
	}
}
