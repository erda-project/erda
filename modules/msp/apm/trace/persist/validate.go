// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package persist

import (
	"errors"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/msp/apm/trace"
)

// Validator .
type Validator interface {
	Validate(s *trace.Span) error
}

type nopValidator struct{}

func (*nopValidator) Validate(*trace.Span) error { return nil }

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

func (v *validator) Validate(s *trace.Span) error {
	if len(s.TraceId) <= 0 {
		return ErrIDEmpty
	}
	return nil
}
