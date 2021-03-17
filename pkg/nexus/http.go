package nexus

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrMissingRepoFormat = errors.New("missing repo format")
)

func (n *Nexus) basicAuthBase64Value() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(n.Username+":"+n.Password))
}

func printJSON(o interface{}) {
	b, _ := json.MarshalIndent(o, "", "  ")
	fmt.Println(string(b))
}

func ErrNotOK(statusCode int, body string) error {
	if statusCode == http.StatusNotFound {
		return ErrNotFound
	}
	return errors.Errorf("status code: %d, err: %v", statusCode, body)
}
