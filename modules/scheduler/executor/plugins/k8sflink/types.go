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

package k8sflink

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/clientgo"
)

type Flink struct {
	Client       *clientgo.ClientSet
	ExecutorName executortypes.Name
	ExecutorKind executortypes.Kind
}

type Option func(f *Flink)

func New(ops ...Option) *Flink {
	f := &Flink{}
	for _, op := range ops {
		op(f)
	}
	return f
}

func WithClient(addr string) Option {
	return func(f *Flink) {
		cs, err := clientgo.New(addr)
		if err != nil {
			logrus.Fatal(err)
		}
		f.Client = cs
	}
}

func WithName(name executortypes.Name) Option {
	return func(f *Flink) {
		f.ExecutorName = name
	}
}

func WithKind(kind executortypes.Kind) Option {
	return func(f *Flink) {
		f.ExecutorKind = kind
	}
}
