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

package persist

import (
	"encoding/json"
	"time"

	"github.com/erda-project/erda-proto-go/oap/entity/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/entity"
)

func (p *provider) decodeData(key, value []byte, topic *string, timestamp time.Time) (interface{}, error) {
	input := &pb.Entity{}
	if err := json.Unmarshal(value, input); err != nil {
		p.stats.DecodeError(value, err)
		if p.Cfg.PrintInvalidData {
			p.Log.Warnf("unknown format entity data: %s", string(value))
		} else {
			p.Log.Warnf("failed to decode entity: %v", err)
		}
		return nil, err
	}
	data := p.normalize(input)
	if err := p.validator.Validate(data); err != nil {
		p.stats.ValidateError(data)
		if p.Cfg.PrintInvalidData {
			p.Log.Warnf("invalid entity data: %s, %s", string(value), err)
		} else {
			p.Log.Warnf("invalid entity: %v", err)
		}
		return nil, err
	}
	return data, nil
}

func (p *provider) handleReadError(err error) error {
	p.Log.Errorf("failed to read entities from kafka: %s", err)
	return nil // return nil to continue read
}

func (p *provider) handleWriteError(list []interface{}, err error) error {
	p.Log.Errorf("failed to write into storage: %s", err)
	return nil // return nil to continue consume
}

func (p *provider) confirmErrorHandler(err error) error {
	p.Log.Errorf("failed to confirm entities from kafka: %s", err)
	return err // return error to exit
}

func (p *provider) normalize(input *pb.Entity) *entity.Entity {
	values := make(map[string]string)
	for k, v := range input.Values {
		values[k] = v.String()
	}
	return &entity.Entity{
		Timestamp:          time.Unix(0, input.CreateTimeUnixNano),
		UpdateTimestamp:    time.Unix(0, input.UpdateTimeUnixNano),
		ID:                 input.Id,
		Type:               input.Type,
		Key:                input.Key,
		Values:             values,
		Labels:             input.Labels,
		CreateTimeUnixNano: input.CreateTimeUnixNano,
		UpdateTimeUnixNano: input.UpdateTimeUnixNano,
	}
}
