package oauth2

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestExtractClientCredentials(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/oauth2/token", nil)
		req.SetBasicAuth("pipeline", "secret")
		id, secret, source := extractClientCredentials(req)
		if id != "pipeline" || secret != "secret" || source != "basic" {
			t.Fatalf("got id=%q secret=%q source=%q", id, secret, source)
		}
	})

	t.Run("query", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/oauth2/token?client_id=pipeline&client_secret=secret", nil)
		id, secret, source := extractClientCredentials(req)
		if id != "pipeline" || secret != "secret" || source != "query" {
			t.Fatalf("got id=%q secret=%q source=%q", id, secret, source)
		}
	})

	t.Run("form", func(t *testing.T) {
		form := url.Values{"client_id": {"pipeline"}, "client_secret": {"secret"}}
		req, _ := http.NewRequest(http.MethodPost, "/oauth2/token", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		id, secret, source := extractClientCredentials(req)
		if id != "pipeline" || secret != "secret" || source != "form" {
			t.Fatalf("got id=%q secret=%q source=%q", id, secret, source)
		}
	})
}
