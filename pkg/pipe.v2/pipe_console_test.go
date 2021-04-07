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

package pipe

//import (
//	"fmt"
//	"testing"
//
//	"github.com/stretchr/testify/require"
//)
//
//func TestPrintStdoutStderr(t *testing.T) {
//	p := Line(
//		ReadFile("/etc/hosts"),
//		System("echo hello 1>&2"),
//		Exec("ping", "-c", "3", "baidu.com"),
//	)
//	outb, errb, err := PrintStdoutStderr(p)
//	require.NoError(t, err)
//	fmt.Println("===========================")
//	fmt.Println(len(outb))
//	fmt.Println(len(errb))
//
//	fmt.Println("===========================stdout")
//	fmt.Println(string(outb))
//
//	fmt.Println("===========================stderr")
//	fmt.Println(string(errb))
//
//	p = Line(
//		ReadFile("/etc/hosts"),
//		Exec("false"),
//	)
//	_, _, err = PrintStdoutStderr(p)
//	require.Error(t, err)
//}
