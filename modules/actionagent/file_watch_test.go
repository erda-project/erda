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

package actionagent

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	f := func(flogs *[]string, line string) {
		*flogs = append(*flogs, line)
	}
	var logs = &[]string{}
	f(logs, "l1")
	fmt.Println(logs)
	f(logs, "l2")
	fmt.Println(logs)
}

func TestMatchStdErrLine(t *testing.T) {
	regexpList := []*regexp.Regexp{}
	regexpStrList := []string{"^[a-z]*can't*"}
	for i := range regexpStrList {
		reg, err := regexp.Compile(regexpStrList[i])
		assert.NoError(t, err)
		regexpList = append(regexpList, reg)
	}
	assert.False(t, matchStdErrLine(regexpList, "failed to open file test.text"))
	assert.True(t, matchStdErrLine(regexpList, "can't open '/jfoejfoijlkjlj/jflejwf.txt': No such file or directory"))
}

func TestDesensitizeLine(t *testing.T) {
	var txtBlackList = []string{"abcdefg", "kskdudh&"}
	tt := []struct {
		line, want string
	}{
		{"admin", "admin"},
		{"abcdefg", "******"},
		{"admin123", "admin123"},
		{"docker login -u abcdefg -p kskdudh&", "docker login -u ****** -p ******"},
	}
	for _, v := range tt {
		desensitizeLine(txtBlackList, &v.line)
		assert.Equal(t, v.want, v.line)
	}

}

func TestAgent_BadWatchFiles(t *testing.T) {
	badAgent := Agent{}
	badAgent.watchFiles()
	assert.True(t, len(badAgent.Errs) > 0)
}

func TestAgent_watchFiles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	agent := Agent{
		Ctx:    ctx,
		Cancel: cancel,
	}
	agent.watchFiles()
	t.Logf("no error here")
}
