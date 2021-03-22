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
