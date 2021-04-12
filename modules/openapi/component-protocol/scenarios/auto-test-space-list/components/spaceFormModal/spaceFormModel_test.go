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

// ^[a-z\u4e00-\u9fa5A-Z0-9_-]*$
package spaceFormModal

//import (
//	"regexp"
//	"testing"
//
//	"gotest.tools/assert"
//)
//
//func TestMatch(t *testing.T) {
//	// str := "1234daS撒_-&*("
//	// reg, _ := regexp.MatchString("[^a-z\u4e00-\u9fa5A-Z0-9_-]", str)
//	// //fmt.Println(reg.FindAllStringSubmatch(str, -1))
//	// assert.Equal(t, reg, true)
//
//	// a := regexp.MustCompile("[^a-z\u4e00-\u9fa5A-Z0-9_-]")
//	// b := regexp.MustCompile("^[a-z\u4e00-\u9fa5A-Z0-9_-]")
//	// assert.Equal(t, a, b)
//	// assert.NoError(t, err)
//	// ^[A-Za-z0-9\u4e00-\u9fa5｜\.｜_｜\-｜ ]+$
//
//	// a := regexp.MustCompile(`^[A-Za-z0-9|\\u4e00-\\u9fa5|\.|_|\-| ]+$`)
//	// a := regexp.MustCompile("^[A-Za-z0-9|\u4e00-\u9fa5|_|-]+$")
//	// matched, err := regexp.MatchString("^[a-zA-Z0-9_.|\\-|\\s|\\u4e00-\\u9fa5]+$", str)
//	// fmt.Println(matched, err)
//	a := regexp.MustCompile("^[a-zA-Z0-9_.|\\-|\\s|\u4e00-\u9fa5]+$")
//	b := a.MatchString("1234daS_.是%……&的-")
//	assert.Equal(t, b, true)
//}
