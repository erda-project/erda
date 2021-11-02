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
	"context"
	"strings"
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
	Begin() int
}

var (
	OutOfLengthSize = errors.New("cmd length out of buffer limit")
)

const (
	startIdx   = 1
	maxBufSize = 2048
)
const (
	// state
	normal = iota
	reverseSearch
	reverseSearched
)

// start with index 1
type commandQueue struct {
	// cmds contains each cmd of history, start from startIdx
	cmds    []string
	length  int
	maxSize int
}

func (c *commandQueue) Front() string {
	return c.cmds[0]
}
func (c *commandQueue) Begin() int {
	return startIdx
}
func (c *commandQueue) pushFront(cmd string) error {
	c.length = mathutil.Min(c.maxSize, c.length+1)
	for i := c.length - 1; i > 0; i-- {
		c.cmds[i+1] = c.cmds[i]
	}
	c.cmds[startIdx] = cmd
	return nil
}

func (c *commandQueue) Back() string {
	return c.cmds[c.length]
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
	if i <= c.length && i >= 0 {
		return c.cmds[i], nil
	}
	return "", errors.Errorf("queue index out of range")
}

type dispatcher struct {
	ctx            context.Context
	closeChan      chan struct{}
	queue          Queue
	cursorLine     int
	cursorIdx      int
	length         int
	buf            []byte
	searchBuf      []byte
	searchIdx      int
	BufferMaxSize  int
	executeCommand string
	auditReqChan   chan *cmdWithTimestamp
	state          int
}

func (d *dispatcher) Print(b byte) error {
	if b == 127 {
		d.state = normal
	}
	switch d.state {
	case reverseSearch:
		return d.reverseSearch(b)
	case normal:
		// DEL
		if b == 127 {
			d.buf = append(d.buf[:d.cursorIdx-1], d.buf[d.cursorIdx:]...)
			d.cursorIdx = mathutil.Max(d.cursorIdx-1, 1)
			d.length = mathutil.Max(d.length-1, 0)
			return nil
		}
		if d.length >= d.BufferMaxSize {
			d.reset()
			return OutOfLengthSize
		}
		for i := d.length; i >= d.cursorIdx; i-- {
			d.buf[i+1] = d.buf[i]
		}
		d.buf[d.cursorIdx] = b
		d.cursorIdx++
		d.length++
	}
	return nil
}

func (d *dispatcher) Execute(b byte) error {
	switch b {
	// ctrl a
	case 1:
		return d.CUB(len(d.buf))
	// ctrl c
	case 3:
		d.state = normal
		d.reset()
		return nil
	case 4:
		d.closeChan <- struct{}{}
	// ctrl e
	case 5:
		return d.CUF(len(d.buf))
	case 13:
		err := d.sendAuditReq()
		if err != nil {
			return err
		}
	// ctrl r
	case 18:
		d.searchIdx = 0
		d.state = reverseSearch
	}
	d.reset()
	recover()
	return nil
}

func (d *dispatcher) sendAuditReq() error {
	d.state = normal
	d.executeCommand = string(d.buf[startIdx : startIdx+d.length])
	if d.executeCommand == "" {
		return nil
	}
	err := d.queue.pushFront(d.executeCommand)
	if err != nil {
		d.reset()
		return err
	}
	d.auditReqChan <- &cmdWithTimestamp{
		start: time.Now(),
		cmd:   d.executeCommand,
	}
	return nil
}

func (d *dispatcher) reset() {
	d.cursorLine = 0
	d.cursorIdx = startIdx
	d.length = 0
	d.executeCommand = ""
}

func (d *dispatcher) reverseSearch(b byte) error {
	d.searchBuf[d.searchIdx] = b
	d.searchIdx++
	if d.searchIdx >= d.BufferMaxSize {
		d.reset()
		return OutOfLengthSize
	}
	searchStr := string(d.searchBuf[:d.searchIdx])
	for i := d.queue.Begin(); i <= d.queue.Size(); i++ {
		str, _ := d.queue.Get(i)
		if strings.Contains(str, searchStr) {
			copy(d.buf[startIdx:], str)
			d.cursorIdx = len(str) + 1
			d.length = len(str)
			d.cursorLine = i
			return nil
		}
	}
	return nil
}

func (d *dispatcher) CUU(i int) error {
	switch d.state {
	case reverseSearch:
		d.state = normal
	}
	d.cursorLine = mathutil.Min(d.queue.Size(), d.cursorLine+i)
	c, err := d.queue.Get(d.cursorLine)
	if err != nil {
		return err
	}
	copy(d.buf[startIdx:], c)
	d.length = len(c)
	d.cursorIdx = len(c)
	return nil
}

func (d *dispatcher) CUD(i int) error {
	switch d.state {
	case reverseSearch:
		d.state = normal
	}
	d.cursorLine = mathutil.Max(0, d.cursorLine-i)
	c, err := d.queue.Get(d.cursorLine)
	if err != nil {
		return err
	}
	copy(d.buf[startIdx:], c)
	d.length = len(c)
	d.cursorIdx = len(c)
	return nil
}

func (d *dispatcher) CUF(i int) error {
	switch d.state {
	case reverseSearch:
		d.state = normal
	}
	d.cursorIdx = mathutil.Min(d.cursorIdx+i, d.length)
	return nil
}

func (d *dispatcher) CUB(i int) error {
	switch d.state {
	case reverseSearch:
		d.state = normal
	}
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

func NewDispatcher(auditReqChan chan *cmdWithTimestamp, closeChan chan struct{}) *dispatcher {
	maxSize := 100

	d := &dispatcher{queue: &commandQueue{
		cmds:    make([]string, maxSize+1),
		length:  0,
		maxSize: maxSize,
	},
		BufferMaxSize: maxBufSize,
		cursorIdx:     startIdx,
		buf:           make([]byte, maxBufSize+1),
		searchBuf:     make([]byte, maxBufSize+1),
		auditReqChan:  auditReqChan,
		closeChan:     closeChan,
	}
	return d
}
