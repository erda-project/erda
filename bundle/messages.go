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
