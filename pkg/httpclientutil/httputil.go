package httpclientutil

import (
	"bytes"
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpclient"

	"github.com/pkg/errors"
)

type RespForRead struct {
	Success bool                      `json:"success"`
	Data    json.RawMessage           `json:"data,omitempty"`
	Err     *apistructs.ErrorResponse `json:"err,omitempty"`
}

func DoJson(r *httpclient.Request, o interface{}) error {
	var b bytes.Buffer
	hr, err := r.Header("Content-Type", "application/json").Do().Body(&b)
	if err != nil {
		return errors.Wrap(err, "failed to request http")
	}
	if !hr.IsOK() {
		return errors.Errorf("failed to request http, status-code %d, content-type %s, raw body %s",
			hr.StatusCode(), hr.ResponseHeader("Content-Type"), b.String())
	}
	var resp RespForRead
	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return errors.Wrapf(err, "response not json, raw body %s", b.String())
	}
	if !resp.Success {
		return errors.Errorf("rest api not success, raw body %s", b.String())
	}
	if o == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(resp.Data), o); err != nil {
		return errors.Wrapf(err, "resp.Data not json, raw string body %s", string(resp.Data))
	}
	return nil
}
