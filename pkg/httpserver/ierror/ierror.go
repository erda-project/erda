package ierror

import (
	"github.com/erda-project/erda/pkg/i18n"
)

type IAPIError interface {
	Render(locale *i18n.LocaleResource) string
	Code() string
	HttpCode() int
}
