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

package console

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/erda-project/erda/modules/core/monitor/log/pb"
)

const (
	Selector = "console"
)

func DefaultDecoderFunc(data []byte) ([]byte, error) {
	lb := &pb.LogBatch{}
	err := lb.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(lb)
}

func New() *Output {
	return &Output{
		DecoderFunc: DefaultDecoderFunc,
		Writer:      os.Stdout,
	}
}

type Decoder func(data []byte) ([]byte, error)

// TODO pretty stdout
type Output struct {
	Writer      io.Writer
	DecoderFunc Decoder
}

func (o *Output) Send(ctx context.Context, data []byte) error {
	buf := make([]byte, 0)
	if o.DecoderFunc != nil {
		tmp, err := o.DecoderFunc(data)
		if err != nil {
			return err
		}
		buf = tmp
	}
	_, err := fmt.Fprintln(o.Writer, string(buf))
	return err
}
