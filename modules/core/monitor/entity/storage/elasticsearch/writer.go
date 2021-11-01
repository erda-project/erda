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

package elasticsearch

import (
	"context"

	"github.com/erda-project/erda-proto-go/oap/entity/pb"
)

type Writer struct {
	p   *provider
	ctx context.Context
}

func (w *Writer) WriteN(vals ...interface{}) (n int, err error) {
	for _, val := range vals {
		e := w.Write(val)
		if e != nil {
			err = e
		} else {
			n++
		}
	}
	return n, err
}

func (w *Writer) Write(val interface{}) error {
	if val == nil {
		return nil
	}
	return w.p.SetEntity(w.ctx, val.(*pb.Entity))
}

func (w *Writer) Close() error { return nil }
