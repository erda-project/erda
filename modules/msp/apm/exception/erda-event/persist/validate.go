package persist

import (
	"errors"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/msp/apm/exception"
)

// Validator .
type Validator interface {
	Validate(s *exception.Erda_event) error
}

type nopValidator struct{}

func (*nopValidator) Validate(*exception.Erda_event) error { return nil }

// NopValidator .
var NopValidator Validator = &nopValidator{}

func newValidator(cfg *config) Validator {
	return &validator{
		bdl: bundle.New(bundle.WithCoreServices(), bundle.WithDOP()),
	}
}

type validator struct {
	bdl *bundle.Bundle
}

var (
	// ErrIDEmpty .
	ErrIDEmpty = errors.New("id empty")
)

func (v *validator) Validate(e *exception.Erda_event) error {
	if len(e.EventId) <= 0 {
		return ErrIDEmpty
	}
	return nil
}
