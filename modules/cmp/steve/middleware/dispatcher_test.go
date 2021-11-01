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
	"reflect"
	"testing"

	"github.com/Azure/go-ansiterm"
)

func TestNewDispatcher(t *testing.T) {
	closeChan := make(chan struct{}, 10)
	auditReqChan := make(chan *cmdWithTimestamp, 10)
	tests := []struct {
		name string
		want *dispatcher
	}{
		{
			name: "1",
			want: NewDispatcher(auditReqChan, closeChan),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDispatcher(auditReqChan, closeChan); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDispatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commandQueue_Back(t *testing.T) {
	type fields struct {
		cmds    []string
		length  int
		maxSize int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "1",
			fields: fields{
				cmds:    make([]string, 10),
				length:  0,
				maxSize: 3,
			},
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commandQueue{
				cmds:    tt.fields.cmds,
				length:  tt.fields.length,
				maxSize: tt.fields.maxSize,
			}
			err := c.pushBack("123")
			if err != nil {
				return
			}
		})
	}
}

func Test_commandQueue_Front(t *testing.T) {
	type fields struct {
		cmds    []string
		length  int
		maxSize int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "1",
			fields: fields{
				cmds:    make([]string, 10),
				length:  0,
				maxSize: 3,
			},
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commandQueue{
				cmds:    tt.fields.cmds,
				length:  tt.fields.length,
				maxSize: tt.fields.maxSize,
			}
			err := c.pushFront("123")
			if err != nil {
				return
			}
		})
	}
}

func Test_commandQueue_Get(t *testing.T) {
	type fields struct {
		cmds    []string
		length  int
		maxSize int
	}
	type args struct {
		i int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			fields: fields{
				cmds:    make([]string, 10),
				length:  0,
				maxSize: 3,
			},
			want: "123",
			args: args{i: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commandQueue{
				cmds:    tt.fields.cmds,
				length:  tt.fields.length,
				maxSize: tt.fields.maxSize,
			}
			c.pushFront("123")
			c.pushFront("123")
			c.pushFront("123")
			got, err := c.Get(tt.args.i)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_commandQueue_Size(t *testing.T) {
	type fields struct {
		cmds    []string
		length  int
		maxSize int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "1",
			fields: fields{
				cmds:    make([]string, 10),
				length:  0,
				maxSize: 3,
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &commandQueue{
				cmds:    tt.fields.cmds,
				length:  tt.fields.length,
				maxSize: tt.fields.maxSize,
			}
			c.pushFront("123")
			if got := c.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dispatcher_Print(t *testing.T) {
	type fields struct {
		queue          Queue
		cursorLine     int
		cursorIdx      int
		length         int
		buf            []byte
		BufferMaxSize  int
		executeCommand string
		cmds           []*cmdWithTimestamp
	}
	closeChan := make(chan struct{}, 10)
	auditReqChan := make(chan *cmdWithTimestamp, 10)
	type args struct {
		b byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{b: 97},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(auditReqChan, closeChan)
			if err := d.Print(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dispatcher_Execute(t *testing.T) {
	type fields struct {
		queue          Queue
		cursorLine     int
		cursorIdx      int
		length         int
		buf            []byte
		BufferMaxSize  int
		executeCommand string
		cmds           []*cmdWithTimestamp
	}
	closeChan := make(chan struct{}, 10)
	auditReqChan := make(chan *cmdWithTimestamp, 10)
	type args struct {
		b byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{b: 13},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(auditReqChan, closeChan)
			if err := d.Execute(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dispatcher_CUU(t *testing.T) {
	type fields struct {
		queue          Queue
		cursorLine     int
		cursorIdx      int
		length         int
		buf            []byte
		BufferMaxSize  int
		executeCommand string
		cmds           []*cmdWithTimestamp
	}
	closeChan := make(chan struct{}, 10)
	auditReqChan := make(chan *cmdWithTimestamp, 10)
	type args struct {
		i int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{i: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(auditReqChan, closeChan)
			parser := ansiterm.CreateParser("Ground", d)
			parser.Parse([]byte{97})
			parser.Parse([]byte{13})
			parser.Parse([]byte{98})
			parser.Parse([]byte{13})
			if err := d.CUU(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("CUU() error = %v, wantErr %v", err, tt.wantErr)
			}
			parser.Parse([]byte{13})
			d.CUU(tt.args.i)
			if err := d.CUD(tt.args.i - 1); (err != nil) != tt.wantErr {
				t.Errorf("CUU() error = %v, wantErr %v", err, tt.wantErr)
			}
			parser.Parse([]byte{13})
		})
	}
}

func Test_dispatcher_CUF(t *testing.T) {
	type fields struct {
		queue          Queue
		cursorLine     int
		cursorIdx      int
		length         int
		buf            []byte
		BufferMaxSize  int
		executeCommand string
		cmds           []*cmdWithTimestamp
	}
	closeChan := make(chan struct{}, 10)
	auditReqChan := make(chan *cmdWithTimestamp, 10)
	type args struct {
		i int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "1",
			args: args{i: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDispatcher(auditReqChan, closeChan)
			parser := ansiterm.CreateParser("Ground", d)
			parser.Parse([]byte{97})
			parser.Parse([]byte{98})
			if err := d.CUF(tt.args.i); (err != nil) != tt.wantErr {
				t.Errorf("CUF() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := d.CUF(tt.args.i - 1); (err != nil) != tt.wantErr {
				t.Errorf("CUF() error = %v, wantErr %v", err, tt.wantErr)
			}
			parser.Parse([]byte{98})
			parser.Parse([]byte{13})
		})
	}
}
