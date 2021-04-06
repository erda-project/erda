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

package loading

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/erda-project/erda/pkg/terminal/color_str"
)

var loadingChars = []string{"⠧", "⠶", "⠼", "⠹", "⠛", "⠏"}
var idx int

func Loading(onelineDesc string, f func(), disappearAfterDone bool, eraseCursor bool) {
	if eraseCursor {
		tput("civis")
		defer tput("cvvis")
	}

	onelineDesc = strings.Replace(onelineDesc, "\n", " ", -1)
	c := make(chan struct{})
	done := false
	f1 := func() {
		f()
		c <- struct{}{}
	}
	go f1()
	printStr := color_str.Green(loadingChars[idx%len(loadingChars)]) + " " + onelineDesc
	fmt.Printf("\r" + printStr)
	idx++
	for !done {
		select {
		case <-c:
			done = true
		case <-time.After(100 * time.Millisecond):
			printStr := color_str.Green(loadingChars[idx%len(loadingChars)]) + " " + onelineDesc
			fmt.Printf("\r" + printStr)
			idx++
		}
	}

	if disappearAfterDone {
		fmt.Printf("\r" + strings.Repeat(" ", len(onelineDesc)+2) + "\r")
	}
}

func tput(arg string) error {
	cmd := exec.Command("tput", arg)
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
