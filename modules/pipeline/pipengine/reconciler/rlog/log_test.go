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

package rlog

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPErrorAndReturn(t *testing.T) {
	err := fmt.Errorf("failed to get schedulable tasks, err: %s", "internal error")
	err = PErrorAndReturn(1, err)
	assert.Error(t, err)
	assert.True(t, true, strings.HasPrefix(err.Error(), pErrorFormat))
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	s := "start watching"
	Infof(s)
	assert.True(t, strings.Contains(buf.String(), fmt.Sprintf(errorFormat, s)))
}
