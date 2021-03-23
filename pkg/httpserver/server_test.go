package httpserver_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/httpserver"
)

func TestServer_Base64EncodeRequestBody(t *testing.T) {
	s := "hello world"
	base64Str := base64.StdEncoding.EncodeToString([]byte(s))

	go func() {
		server := httpserver.New(":8080")
		server.RegisterEndpoint([]httpserver.Endpoint{
			{
				Path:   "/base64-test",
				Method: http.MethodPost,
				Handler: func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
					b, err := ioutil.ReadAll(r.Body)
					if err != nil {
						return nil, err
					}
					return httpserver.OkResp(string(b) == s)
				},
			},
		})
		server.ListenAndServe()
	}()

	time.Sleep(time.Second)

	// without base64 header
	resp, err := http.DefaultClient.Post("http://localhost:8080/base64-test", httpserver.ContentTypeJSON, bytes.NewBufferString(base64Str))
	assert.NoError(t, err)
	var result httpserver.Resp
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.False(t, result.Data.(bool))

	// with base64 header
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/base64-test", bytes.NewBufferString(base64Str))
	assert.NoError(t, err)
	req.Header["Content-Type"] = []string{httpserver.ContentTypeJSON}
	req.Header[httpserver.Base64EncodedRequestBody] = []string{"true"}
	resp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)
	assert.True(t, result.Data.(bool))
}
