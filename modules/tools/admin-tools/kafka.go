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

package admin_tools

import (
	"strings"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/kafka"
)

func (p *provider) showKafkaChannelSize() interface{} {
	return api.Success(map[string]interface{}{
		"produce_pending": p.kafka.ProduceChannelSize(),
		"events_pending":  p.kafka.ProduceEventsChannelSize(),
	})
}

func (p *provider) pushKafkaData(param struct {
	Topics  string `query:"topics" json:"topics"`
	Payload string `query:"payload" json:"payload"`
}) interface{} {
	param.Topics = strings.TrimSpace(param.Topics)
	if len(param.Topics) < 0 {
		return api.Errors.InvalidParameter("topics must not be empty")
	}
	if len(param.Payload) < 0 {
		return api.Errors.InvalidParameter("payload must not be empty")
	}
	for _, topic := range strings.Split(param.Topics, ",") {
		t := strings.TrimSpace(topic)
		if len(topic) <= 0 {
			continue
		}
		err := p.kafka.producer.Write(&kafka.Message{
			Topic: &t,
			Data:  []byte(param.Payload),
		})
		if err != nil {
			return api.Errors.Internal(err)
		}
	}
	return api.Success("OK")
}
