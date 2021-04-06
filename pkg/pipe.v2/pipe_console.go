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

import (
	"io"
	"os"
)

// OutputWithPrintStderr runs the p pipe and returns its stdout output,
// and print stderr to console.
//
// See functions Output.
func OutputWithPrintStderr(p Pipe) ([]byte, error) {
	outb := &OutputBuffer{}
	s := NewState(outb, os.Stderr)
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), err
}

func PrintStdoutStderr(p Pipe) ([]byte, []byte, error) {
	outb := &OutputBuffer{}
	errb := &OutputBuffer{}
	s := NewState(io.MultiWriter(os.Stdout, outb), io.MultiWriter(os.Stderr, errb))
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), errb.Bytes(), err
}
