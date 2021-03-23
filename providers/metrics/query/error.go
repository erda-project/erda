package query

import (
	"encoding/json"
	"fmt"
)

type returnedErrorFormat struct {
	Success string            `json:"success"`
	Err     map[string]string `json:"err"`
}

type ServerError struct {
	errorCode string
	message   string
	context   string
}

func NewServerError(body []byte) error {
	data := &returnedErrorFormat{}
	if err := json.Unmarshal(body, data); err != nil {
		return err
	}
	return &ServerError{
		errorCode: data.Err["code"],
		message:   data.Err["msg"],
		context:   data.Err["context"],
	}
}

func (e ServerError) Error() string {
	return fmt.Sprintf("SDK.ServerError:\nErrorCode: %s\nContext: %s\nMessage: %s", e.errorCode, e.context, e.message)
}

