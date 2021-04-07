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

package dumpstack

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func Open() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	go func() {
		for range c {
			dumpStacks()
		}
	}()
}

func dumpStacks() {
	var (
		buf  []byte
		size int
	)

	bufLen := 16 * 1024 // 16K

	for size == len(buf) {
		buf = make([]byte, bufLen)
		size = runtime.Stack(buf, true)
		bufLen *= 2
	}
	buf = buf[:size]

	fmt.Printf("=== BEGIN goroutine stack === \n%s\n=== END goroutine stack ===\n", buf)
}
