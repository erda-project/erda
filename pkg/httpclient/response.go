package httpclient

import (
	"net/http"
)

// Response 定义 http 应答对象.
type Response struct {
	body     []byte
	internal *http.Response
}

// StatusCode return http status code.
func (r *Response) StatusCode() int {
	return r.internal.StatusCode
}

// IsOK 返回 200 与否.
func (r *Response) IsOK() bool {
	return r.StatusCode()/100 == 2
}

// IsNotfound 返回 404 与否.
func (r *Response) IsNotfound() bool {
	return r.StatusCode() == http.StatusNotFound
}

// IsConflict 返回 409 与否.
func (r *Response) IsConflict() bool {
	return r.StatusCode() == http.StatusConflict
}

// IsBadRequest 返回 400.
func (r *Response) IsBadRequest() bool {
	return r.StatusCode() == http.StatusBadRequest
}

// ResponseHeader 返回指定应答 header 值.
func (r *Response) ResponseHeader(key string) string {
	return r.internal.Header.Get(key)
}

// Headers 返回resp的header信息
func (r *Response) Headers() http.Header {
	return r.internal.Header
}

func (r *Response) Body() []byte {
	return r.body
}
