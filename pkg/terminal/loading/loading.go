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
