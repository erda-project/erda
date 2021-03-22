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
