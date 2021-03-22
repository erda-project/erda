// pipe - Unix-like pipelines for Go
//
// Copyright (c) 2013 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice, this
//    list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright notice,
//    this list of conditions and the following disclaimer in the documentation
//    and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
// ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// Package pipe implements unix-like pipelines for Go.
//
// See the documentation for details:
//
//   http://labix.org/pipe
//
package pipe

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Pipe functions implement arbitrary functionality that may be
// integrated with pipe scripts and pipelines. Pipe functions
// must not block reading or writing to the state streams. These
// operations must be run from a Task.
type Pipe func(s *State) error

// A Task may be registered by a Pipe into a State to run any
// activity concurrently with other tasks.
// PipelineTasks registered within the execution of a Script only run after
// the preceding entries in the same script have succeeded.
type Task interface {

	// Run runs the task concurrently with other tasks as appropriate
	// for the pipe. Run may flow data from the input stream and/or
	// to the output streams of the pipe, and it must block while doing
	// so. It must return only after all of its activities have
	// terminated completely.
	Run(s *State) error

	// Kill abruptly interrupts in-progress activities being done by Run.
	// If Run is blocked simply reading from and/or writing to the state
	// streams, Kill doesn't have to do anything as Run will be unblocked
	// by the closing of the streams.
	Kill()
}

// State defines the environment for Pipe functions to run on.
// Create a new State via the NewState function.
type State struct {

	// Stdin, Stdout, and Stderr represent the respective data streams
	// that the Pipe may act upon. Reading from and/or writing to these
	// streams must be done from within a Task registered via AddTask.
	// The three streams are initialized by NewState and must not be nil.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// Dir represents the directory in which all filesystem-related
	// operations performed by the Pipe must be run on. It defaults
	// to the current directory, and may be changed by Pipe functions.
	Dir string

	// Env is the process environment in which all executions performed
	// by the Pipe must be run on. It defaults to a copy of the
	// environmnet from the current process, and may be changed by Pipe
	// functions.
	Env []string

	// Timeout defines the amount of time to wait before aborting running tasks.
	// If set to zero, the pipe will not be aborted.
	Timeout time.Duration

	killedMutex sync.Mutex
	killedNoted bool
	killed      chan bool

	pendingTasks []*pendingTask
}

// NewState returns a new state for running pipes with.
// The state's Stdout and Stderr are set to the provided streams,
// Stdin is initialized to an empty reader, and Env is initialized to
// the environment of the current process.
func NewState(stdout, stderr io.Writer) *State {
	if stdout == nil {
		stdout = ioutil.Discard
	}
	if stderr == nil {
		stderr = ioutil.Discard
	}
	return &State{
		Stdin:  strings.NewReader(""),
		Stdout: stdout,
		Stderr: stderr,
		Env:    os.Environ(),
		killed: make(chan bool, 1),
	}
}

type pendingTask struct {
	s State
	t Task
	c []io.Closer

	wg sync.WaitGroup
	wt []*pendingTask

	cancel int32
}

func (pt *pendingTask) closeWhenDone(c io.Closer) {
	pt.c = append(pt.c, c)
}

func (pt *pendingTask) waitFor(other *pendingTask) {
	pt.wg.Add(1)
	other.wt = append(other.wt, pt)
}

func (pt *pendingTask) wait() {
	pt.wg.Wait()
}

func (pt *pendingTask) done(err error) {
	for _, c := range pt.c {
		c.Close()
	}
	for _, wt := range pt.wt {
		if err != nil {
			atomic.AddInt32(&wt.cancel, 1)
		}
		wt.wg.Done()
	}
}

var (
	ErrTimeout = errors.New("timeout")
	ErrKilled  = errors.New("explicitly killed")
)

type Errors []error

func (e Errors) Error() string {
	var errors []string
	for _, err := range e {
		errors = append(errors, err.Error())
	}
	return strings.Join(errors, "; ")
}

// AddTask adds t to be run concurrently with other tasks
// as appropriate for the pipe.
func (s *State) AddTask(t Task) error {
	pt := &pendingTask{s: *s, t: t}
	pt.s.Env = append([]string(nil), s.Env...)
	s.pendingTasks = append(s.pendingTasks, pt)
	return nil
}

// RunTasks runs all pending tasks registered via AddTask.
// This is called by the pipe running functions and generally
// there's no reason to call it directly.
func (s *State) RunTasks() error {
	done := make(chan error, len(s.pendingTasks))
	for _, f := range s.pendingTasks {
		go func(pt *pendingTask) {
			pt.wait()
			var err error
			if pt.cancel == 0 {
				err = pt.t.Run(&pt.s)
			}
			pt.done(err)
			done <- err
		}(f)
	}

	var timeout <-chan time.Time
	if s.Timeout > 0 {
		timeout = time.After(s.Timeout)
	}

	var errs Errors
	var goodErr, badErr bool

	fail := func(err error) {
		if errs == nil {
			for _, pt := range s.pendingTasks {
				pt.t.Kill()
			}
		}
		if errs == nil || errs[len(errs)-1] != ErrTimeout && errs[len(errs)-1] != ErrKilled {
			errs = append(errs, err)
			if discardErr(err) {
				badErr = true
			} else {
				goodErr = true
			}
		}
	}

	for range s.pendingTasks {
		var err error
		select {
		case err = <-done:
		case <-timeout:
			fail(ErrTimeout)
			err = <-done
		case <-s.killed:
			fail(ErrKilled)
			err = <-done
		}
		if err != nil {
			fail(err)
		}
	}
	s.pendingTasks = nil

	if errs == nil {
		return nil
	}

	if goodErr && badErr {
		good := 0
		for _, err := range errs {
			if !discardErr(err) {
				errs[good] = err
				good++
			}
		}
		errs = errs[:good]
	}
	return errs
}

func discardErr(err error) bool {
	if err == io.ErrClosedPipe {
		return true
	}
	if err1, ok := err.(*execError); ok {
		if err2, ok := err1.err.(*exec.ExitError); ok {
			status, ok := err2.Sys().(syscall.WaitStatus)
			return ok && status.Signaled() && status.Signal() == 9
		}
	}
	return false
}

// Kill sends a kill notice to all pending tasks.
func (s *State) Kill() {
	s.killedMutex.Lock()
	if !s.killedNoted {
		s.killedNoted = true
		s.killed <- true
	}
	s.killedMutex.Unlock()
}

// EnvVar returns the value for the named environment variable in s.
func (s *State) EnvVar(name string) string {
	prefix := name + "="
	for _, kv := range s.Env {
		if strings.HasPrefix(kv, prefix) {
			return kv[len(prefix):]
		}
	}
	return ""
}

// SetEnvVar sets the named environment variable to the given value in s.
func (s *State) SetEnvVar(name, value string) {
	prefix := name + "="
	for i, kv := range s.Env {
		if strings.HasPrefix(kv, prefix) {
			s.Env[i] = prefix + value
			return
		}
	}
	s.Env = append(s.Env, prefix+value)
}

// Path returns the provided path relative to the state's current directory.
// If multiple arguments are provided, they're joined via filepath.Join.
// If path is absolute, it is taken by itself.
func (s *State) Path(path ...string) string {
	if len(path) == 0 {
		return s.Dir
	}
	if filepath.IsAbs(path[0]) {
		return filepath.Join(path...)
	}
	if len(path) == 1 {
		return filepath.Join(s.Dir, path[0])
	}
	return filepath.Join(append([]string{s.Dir}, path...)...)
}

func firstErr(err1, err2 error) error {
	if err1 != nil {
		return err1
	}
	return err2
}

// Run runs the p pipe discarding its output.
//
// See functions Output, CombinedOutput, and DividedOutput.
func Run(p Pipe) error {
	s := NewState(nil, nil)
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return err
}

// RunTimeout runs the p pipe discarding its output.
//
// The pipe is killed if it takes longer to run than the provided timeout.
//
// See functions OutputTimeout, CombinedOutputTimeout, and DividedOutputTimeout.
func RunTimeout(p Pipe, timeout time.Duration) error {
	s := NewState(nil, nil)
	s.Timeout = timeout
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return err
}

// Output runs the p pipe and returns its stdout output.
//
// See functions Run, CombinedOutput, and DividedOutput.
func Output(p Pipe) ([]byte, error) {
	outb := &OutputBuffer{}
	s := NewState(outb, nil)
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), err
}

// OutputTimeout runs the p pipe and returns its stdout output.
//
// The pipe is killed if it takes longer to run than the provided timeout.
//
// See functions RunTimeout, CombinedOutputTimeout, and DividedOutputTimeout.
func OutputTimeout(p Pipe, timeout time.Duration) ([]byte, error) {
	outb := &OutputBuffer{}
	s := NewState(outb, nil)
	s.Timeout = timeout
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), err
}

// CombinedOutput runs the p pipe and returns its stdout and stderr
// outputs merged together.
//
// See functions Run, Output, and DividedOutput.
func CombinedOutput(p Pipe) ([]byte, error) {
	outb := &OutputBuffer{}
	s := NewState(outb, outb)
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), err
}

// CombinedOutputTimeout runs the p pipe and returns its stdout and stderr
// outputs merged together.
//
// The pipe is killed if it takes longer to run than the provided timeout.
//
// See functions RunTimeout, OutputTimeout, and DividedOutputTimeout.
func CombinedOutputTimeout(p Pipe, timeout time.Duration) ([]byte, error) {
	outb := &OutputBuffer{}
	s := NewState(outb, outb)
	s.Timeout = timeout
	err := p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), err
}

// DividedOutput runs the p pipe and returns its stdout and stderr outputs.
//
// See functions Run, Output, and CombinedOutput.
func DividedOutput(p Pipe) (stdout []byte, stderr []byte, err error) {
	outb := &OutputBuffer{}
	errb := &OutputBuffer{}
	s := NewState(outb, errb)
	err = p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), errb.Bytes(), err
}

// DividedOutputTimeout runs the p pipe and returns its stdout and stderr outputs.
//
// The pipe is killed if it takes longer to run than the provided timeout.
//
// See functions RunTimeout, OutputTimeout, and CombinedOutputTimeout.
func DividedOutputTimeout(p Pipe, timeout time.Duration) (stdout []byte, stderr []byte, err error) {
	outb := &OutputBuffer{}
	errb := &OutputBuffer{}
	s := NewState(outb, errb)
	s.Timeout = timeout
	err = p(s)
	if err == nil {
		err = s.RunTasks()
	}
	return outb.Bytes(), errb.Bytes(), err
}

// OutputBuffer is a concurrency safe writer that buffers all input.
//
// It is used in the implementation of the output functions.
type OutputBuffer struct {
	m   sync.Mutex
	buf []byte
}

// Writes appends b to out's buffered data.
func (out *OutputBuffer) Write(b []byte) (n int, err error) {
	out.m.Lock()
	out.buf = append(out.buf, b...)
	out.m.Unlock()
	return len(b), nil
}

// Bytes returns all the data written to out.
func (out *OutputBuffer) Bytes() []byte {
	out.m.Lock()
	buf := out.buf
	out.m.Unlock()
	return buf
}

// Exec returns a pipe that runs the named program with the given arguments.
func Exec(name string, args ...string) Pipe {
	return func(s *State) error {
		s.AddTask(&execTask{name: name, args: args})
		return nil
	}
}

// System returns a pipe that runs cmd via a system shell.
// It is equivalent to the pipe Exec("/bin/sh", "-c", cmd).
func System(cmd string) Pipe {
	return Exec("/bin/sh", "-c", cmd)
}

type execTask struct {
	name string
	args []string

	m      sync.Mutex
	p      *os.Process
	cancel bool
}

func (f *execTask) Run(s *State) error {
	f.m.Lock()
	if f.cancel {
		f.m.Unlock()
		return nil
	}
	cmd := exec.Command(f.name, f.args...)
	cmd.Dir = s.Dir
	cmd.Env = s.Env
	cmd.Stdin = s.Stdin
	cmd.Stdout = s.Stdout
	cmd.Stderr = s.Stderr
	err := cmd.Start()
	f.p = cmd.Process
	f.m.Unlock()
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return &execError{f.name, err}
	}
	return nil
}

func (f *execTask) Kill() {
	f.m.Lock()
	p := f.p
	f.cancel = true
	f.m.Unlock()
	if p != nil {
		p.Kill()
	}
}

type execError struct {
	name string
	err  error
}

func (e *execError) Error() string {
	return fmt.Sprintf("command %q: %v", e.name, e.err)
}

// ChDir changes the pipe's current directory. If dir is relative,
// the change is made relative to the pipe's previous current directory.
//
// Other than it being the default current directory for new pipes,
// the working directory of the running process isn't considered or
// changed.
func ChDir(dir string) Pipe {
	return func(s *State) error {
		s.Dir = s.Path(dir)
		return nil
	}
}

// MkDir creates dir with the provided perm bits. If dir is relative,
// the created path is relative to the pipe's current directory.
func MkDir(dir string, perm os.FileMode) Pipe {
	return func(s *State) error {
		return os.Mkdir(s.Path(dir), perm)
	}
}

// MkDirAll creates the missing parents of dir and dir itself with the
// provided perm bits. If dir is relative, the created path is relative
// to the pipe's current directory.
func MkDirAll(dir string, perm os.FileMode) Pipe {
	return func(s *State) error {
		return os.MkdirAll(s.Path(dir), perm)
	}
}

// SetEnvVar sets the value of the named environment variable in the pipe.
//
// Other than it being the default for new pipes, the environment of the
// running process isn't consulted or changed.
func SetEnvVar(name string, value string) Pipe {
	return func(s *State) error {
		s.SetEnvVar(name, value)
		return nil
	}
}

// Line creates a pipeline with the provided entries, where the stdout
// of entry N in the pipeline is connected to the stdin of entry N+1.
//
// For example, the equivalent of "cat article.ps | lpr" is:
//
//    p := pipe.Line(
//        pipe.ReadFile("article.ps"),
//        pipe.Exec("lpr"),
//    )
//    output, err := pipe.CombinedOutput(p)
//
func Line(p ...Pipe) Pipe {
	return func(s *State) error {
		dir := s.Dir
		env := s.Env
		s.Env = append([]string(nil), s.Env...)
		defer func() {
			s.Dir = dir
			s.Env = env
		}()

		end := len(p) - 1
		endStdout := s.Stdout
		var r *io.PipeReader
		var w *io.PipeWriter
		for i, p := range p {
			var closeIn, closeOut *refCloser
			if r != nil {
				closeIn = &refCloser{r, 1}
			}
			if i == end {
				r, w = nil, nil
				s.Stdout = endStdout
			} else {
				r, w = io.Pipe()
				s.Stdout = w
				closeOut = &refCloser{w, 1}
			}

			oldLen := len(s.pendingTasks)
			if err := p(s); err != nil {
				closeIn.Close()
				return err
			}
			newLen := len(s.pendingTasks)

			for fi := oldLen; fi < newLen; fi++ {
				pt := s.pendingTasks[fi]
				if c, ok := pt.s.Stdin.(io.Closer); ok && closeIn.uses(c) {
					closeIn.refs++
					pt.closeWhenDone(closeIn)
				}
				if c, ok := pt.s.Stdout.(io.Closer); ok && closeOut.uses(c) {
					closeOut.refs++
					pt.closeWhenDone(closeOut)
				}
				if c, ok := pt.s.Stderr.(io.Closer); ok && closeOut.uses(c) {
					closeOut.refs++
					pt.closeWhenDone(closeOut)
				}
			}
			closeIn.Close()
			closeOut.Close()

			if i < end {
				s.Stdin = r
			}
		}
		return nil
	}
}

type refCloser struct {
	c    io.Closer
	refs int32
}

func (rc *refCloser) uses(c io.Closer) bool {
	return rc != nil && rc.c == c
}

func (rc *refCloser) Close() error {
	if rc != nil && atomic.AddInt32(&rc.refs, -1) == 0 {
		return rc.c.Close()
	}
	return nil
}

// Script creates a pipe sequence with the provided entries.
//
// For example, the equivalent of "cat article.ps | lpr; mv article.ps{,.done}" is:
//
//    p := pipe.Script(
//        pipe.Line(
//            pipe.ReadFile("article.ps"),
//            pipe.Exec("lpr"),
//        ),
//        pipe.RenameFile("article.ps", "article.ps.done"),
//    )
//    output, err := pipe.CombinedOutput(p)
//
func Script(p ...Pipe) Pipe {
	return func(s *State) error {
		saved := *s
		s.Env = append([]string(nil), s.Env...)
		defer func() {
			s.Dir = saved.Dir
			s.Env = saved.Env
		}()

		startLen := len(s.pendingTasks)
		for _, p := range p {
			oldLen := len(s.pendingTasks)
			if err := p(s); err != nil {
				return err
			}
			newLen := len(s.pendingTasks)

			s.Stdin = saved.Stdin
			s.Stdout = saved.Stdout
			s.Stderr = saved.Stderr

			for fi := oldLen; fi < newLen; fi++ {
				for wi := startLen; wi < oldLen; wi++ {
					s.pendingTasks[fi].waitFor(s.pendingTasks[wi])
				}
			}
		}
		return nil
	}
}

type taskFunc func(s *State) error

func (f taskFunc) Run(s *State) error { return f(s) }
func (f taskFunc) Kill()              {}

// TaskFunc is a helper to define a Pipe that adds a Task
// with f as its Run method.
func TaskFunc(f func(s *State) error) Pipe {
	return func(s *State) error {
		s.AddTask(taskFunc(f))
		return nil
	}
}

// Print provides args to fmt.Sprint and writes the resuling
// string to the pipe's stdout.
func Print(args ...interface{}) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := s.Stdout.Write([]byte(fmt.Sprint(args...)))
		return err
	})
}

// Println provides args to fmt.Sprintln and writes the resuling
// string to the pipe's stdout.
func Println(args ...interface{}) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := s.Stdout.Write([]byte(fmt.Sprintln(args...)))
		return err
	})
}

// Printf provides format and args to fmt.Sprintf and writes
// the resulting string to the pipe's stdout.
func Printf(format string, args ...interface{}) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := s.Stdout.Write([]byte(fmt.Sprintf(format, args...)))
		return err
	})
}

// Read reads data from r and writes it to the pipe's stdout.
func Read(r io.Reader) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := io.Copy(s.Stdout, r)
		return err
	})
}

// Write writes to w the data read from the pipe's stdin.
func Write(w io.Writer) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := io.Copy(w, s.Stdin)
		return err
	})
}

// Discard reads data from the pipe's stdin and discards it.
func Discard() Pipe {
	return Write(ioutil.Discard)
}

// Tee reads data from the pipe's stdin and writes it both to
// the pipe's stdout and to w.
func Tee(w io.Writer) Pipe {
	return TaskFunc(func(s *State) error {
		_, err := io.Copy(w, io.TeeReader(s.Stdin, s.Stdout))
		return err
	})
}

// ReadFile reads data from the file at path and writes it to the
// pipe's stdout.
func ReadFile(path string) Pipe {
	return TaskFunc(func(s *State) error {
		file, err := os.Open(s.Path(path))
		if err != nil {
			return err
		}
		_, err = io.Copy(s.Stdout, file)
		file.Close()
		return err
	})
}

// WriteFile writes to the file at path the data read from the
// pipe's stdin. If the file doesn't exist, it is created with perm.
func WriteFile(path string, perm os.FileMode) Pipe {
	return TaskFunc(func(s *State) error {
		file, err := os.OpenFile(s.Path(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, s.Stdin)
		return firstErr(err, file.Close())
	})
}

// AppendFile append to the end of the file at path the data read
// from the pipe's stdin. If the file doesn't exist, it is created
// with perm.
func AppendFile(path string, perm os.FileMode) Pipe {
	return TaskFunc(func(s *State) error {
		file, err := os.OpenFile(s.Path(path), os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, s.Stdin)
		return firstErr(err, file.Close())
	})
}

// TeeWriteFile reads data from the pipe's stdin and writes it both to
// the pipe's stdout and to the file at path. If the file doesn't
// exist, it is created with perm.
func TeeWriteFile(path string, perm os.FileMode) Pipe {
	return TaskFunc(func(s *State) error {
		file, err := os.OpenFile(s.Path(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, io.TeeReader(s.Stdin, s.Stdout))
		return firstErr(err, file.Close())
	})
}

// TeeAppendFile reads data from the pipe's stdin and writes it both to
// the pipe's stdout and to the file at path. If the file doesn't
// exist, it is created with perm.
func TeeAppendFile(path string, perm os.FileMode) Pipe {
	return TaskFunc(func(s *State) error {
		file, err := os.OpenFile(s.Path(path), os.O_WRONLY|os.O_CREATE|os.O_APPEND, perm)
		if err != nil {
			return err
		}
		_, err = io.Copy(file, io.TeeReader(s.Stdin, s.Stdout))
		return firstErr(err, file.Close())
	})
}

// Replace filters lines read from the pipe's stdin and writes
// the returned values to stdout.
func Replace(f func(line []byte) []byte) Pipe {
	return TaskFunc(func(s *State) error {
		r := bufio.NewReader(s.Stdin)
		for {
			line, err := r.ReadBytes('\n')
			if len(line) > 0 {
				line := f(line)
				if len(line) > 0 {
					_, err := s.Stdout.Write(line)
					if err != nil {
						return err
					}
				}
			}
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}
		panic("unreachable")
	})
}

// Filter filters lines read from the pipe's stdin so that only those
// for which f is true are written to the pipe's stdout.
// The line provided to f has '\n' and '\r' trimmed.
func Filter(f func(line []byte) bool) Pipe {
	return Replace(func(line []byte) []byte {
		if f(bytes.TrimRight(line, "\r\n")) {
			return line
		}
		return nil
	})
}

// RenameFile renames the file fromPath as toPath.
func RenameFile(fromPath, toPath string) Pipe {
	// Register it as a task function so that within scripts
	// it holds until all the preceding flushing is done.
	return TaskFunc(func(s *State) error {
		return os.Rename(s.Path(fromPath), s.Path(toPath))
	})
}
