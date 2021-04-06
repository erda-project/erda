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

package bundle

import (
	"bytes"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

// CreateMessage 发送消息（不包括 event ）, `Message.Labels` 见 `MessageLabel`
func (b *Bundle) CreateMessage(message *apistructs.MessageCreateRequest) error {
	host, err := b.urls.EventBox()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Post(host).Path("/api/dice/eventbox/message/create").
		Header("Accept", "application/json").
		JSONBody(&message).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to create message, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}
