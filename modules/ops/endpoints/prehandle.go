package endpoints

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

type endpoint func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error)

func auth(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		// TODO: auth
		return f(ctx, r, vars)
	}
}

func i18nPrinter(f endpoint) endpoint {
	return func(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
		lang := r.Header.Get("Lang")
		p := message.NewPrinter(language.English)
		if strutil.Contains(lang, "zh") {
			p = message.NewPrinter(language.SimplifiedChinese)
		}
		ctx2 := context.WithValue(ctx, "i18nPrinter", p)
		return f(ctx2, r, vars)
	}
}
