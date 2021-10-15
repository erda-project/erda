// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package middleware

import (
	"time"

	"github.com/pkg/errors"
	"modernc.org/mathutil"
)

type Queue interface {
	Front() string
	pushFront(cmd string) error
	Back() string
	pushBack(cmd string) error
	Size() int
	Get(i int) (string, error)
}

var (
	OutOfLengthSize = errors.New("cmd length out of buffer limit")
)

const (
	startIdx   = 1
	maxBufSize = 2048
)

type commandQueue struct {
	cmds    []string
	length  int
	maxSize int
}

func (c *commandQueue) Front() string {
	return c.cmds[0]
}

func (c *commandQueue) pushFront(cmd string) error {
	c.length = mathutil.Min(c.maxSize, c.length+1)
	for i := c.length - 2; i >= 0; i-- {
		c.cmds[i+1] = c.cmds[i]
	}
	c.cmds[0] = cmd
	return nil
}

func (c *commandQueue) Back() string {
	return c.cmds[c.length-1]
}

func (c *commandQueue) pushBack(cmd string) error {
	c.length = mathutil.Min(c.maxSize, c.length+1)
	c.cmds[c.length-1] = cmd
	return nil
}

func (c *commandQueue) Size() int {
	return c.length
}

func (c *commandQueue) Get(i int) (string, error) {
	if i < c.length && i >= 0 {
		return c.cmds[i], nil
	}
	return "", errors.Errorf("queue index out of range")
}

type dispatcher struct {
	queue          Queue
	cursorLine     int
	cursorIdx      int
	length         int
	buf            []byte
	BufferMaxSize  int
	executeCommand string
	cmds           []*cmdWithTimestamp
}

func (d *dispatcher) Print(b byte) error {
	if b == 127 {
		d.buf = append(d.buf[:d.cursorIdx-1], d.buf[d.cursorIdx:]...)
		d.cursorIdx = mathutil.Max(d.cursorIdx-1, 1)
		d.length = mathutil.Max(d.length-1, 0)
		return nil
	}
	if d.cursorIdx >= d.BufferMaxSize {
		return OutOfLengthSize
	}
	for i := d.length; i >= d.cursorIdx; i-- {
		d.buf[i+1] = d.buf[i]
	}
	d.buf[d.cursorIdx] = b
	d.cursorIdx++
	d.length++
	return nil
}

func (d *dispatcher) Execute(b byte) error {
	d.executeCommand = string(d.buf[startIdx : startIdx+d.length])
	switch b {
	case 1:
		return d.CUB(len(d.buf))
	case 5:
		return d.CUF(len(d.buf))
	}
	if d.executeCommand == "" {
		return nil
	}
	err := d.queue.pushFront(d.executeCommand)
	if err != nil {
		return err
	}
	d.cursorLine = -1
	d.cursorIdx = 1
	d.length = 0
	d.cmds = append(d.cmds, &cmdWithTimestamp{
		start: time.Now(),
		cmd:   d.executeCommand,
	})
	d.executeCommand = ""
	recover()
	return nil
}

func (d *dispatcher) CUU(i int) error {
	c, err := d.queue.Get(d.cursorLine + i)
	if err != nil {
		return err
	}
	d.cursorLine += i
	for i := 0; i < len(c); i++ {
		d.buf[startIdx+i] = c[i]
	}
	d.length = len(c)
	d.cursorIdx = len(c)
	return nil
}

func (d *dispatcher) CUD(i int) error {
	c, err := d.queue.Get(d.cursorLine - i)
	if err != nil {
		return err
	}
	d.cursorLine -= i
	for i := 0; i < len(c); i++ {
		d.buf[startIdx+i] = c[i]
	}
	d.length = len(c)
	d.cursorIdx = len(c)
	return nil
}

func (d *dispatcher) CUF(i int) error {
	d.cursorIdx = mathutil.Min(d.cursorIdx+i, d.length)
	return nil
}

func (d *dispatcher) CUB(i int) error {
	d.cursorIdx = mathutil.Max(d.cursorIdx-i, 1)
	return nil
}

func (d *dispatcher) CNL(i int) error {
	return nil
}

func (d *dispatcher) CPL(i int) error {
	return nil
}

func (d *dispatcher) CHA(i int) error {
	return nil
}

func (d *dispatcher) VPA(i int) error {
	return nil
}

func (d *dispatcher) CUP(i int, i2 int) error {
	return nil
}

func (d *dispatcher) HVP(i int, i2 int) error {
	return nil
}

func (d *dispatcher) DECTCEM(b bool) error {
	return nil
}

func (d *dispatcher) DECOM(b bool) error {
	return nil
}

func (d *dispatcher) DECCOLM(b bool) error {
	return nil
}

func (d *dispatcher) ED(i int) error {
	return nil
}

func (d *dispatcher) EL(i int) error {
	return nil
}

func (d *dispatcher) IL(i int) error {
	return nil
}

func (d *dispatcher) DL(i int) error {
	return nil
}

func (d *dispatcher) ICH(i int) error {
	return nil
}

func (d *dispatcher) DCH(i int) error {
	return nil
}

func (d *dispatcher) SGR(ints []int) error {
	return nil
}

func (d *dispatcher) SU(i int) error {
	return nil
}

func (d *dispatcher) SD(i int) error {
	return nil
}

func (d *dispatcher) DA(strings []string) error {
	return nil
}

func (d *dispatcher) DECSTBM(i int, i2 int) error {
	return nil
}

func (d *dispatcher) IND() error {
	return nil
}

func (d *dispatcher) RI() error {
	return nil
}

func (d *dispatcher) Flush() error {
	return nil
}

func NewDispatcher() *dispatcher {
	maxSize := 100
	return &dispatcher{queue: &commandQueue{
		cmds:    make([]string, maxSize+1),
		length:  0,
		maxSize: maxSize,
	},
		BufferMaxSize: maxBufSize,
		cursorIdx:     1,
		buf:           make([]byte, maxBufSize+1),
	}
}
