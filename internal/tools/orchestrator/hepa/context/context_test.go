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

package context_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	context1 "github.com/erda-project/erda/internal/tools/orchestrator/hepa/context"
)

func Func1(ctx context1.LogContext) {
	ctx.Entry().WithField("tttt", "tttt").Infoln("this is Func1")
}

func Func2(ctx context1.LogContext) {
	ctx.Entry().Infoln("this is Func2")
}

func Func3(ctx context1.LogContext) {
	ctx.Entry().Infoln("this is Func3")
	Func1(ctx)
	Func2(ctx)
}

func TestLogContext_Log(t *testing.T) {
	ctx := context1.WithLoggerIfWithout(context.Background(), logrus.StandardLogger())
	ctx.Entry().Infoln("test")
	Func1(*ctx)
	Func2(*ctx)
	Func3(*ctx)
}

// need not test
func TestWithEntryIfWithout(t *testing.T) {
	ctx := context1.WithEntryIfWithout(context.Background(), logrus.NewEntry(logrus.StandardLogger()))
	context1.WithEntryIfWithout(ctx, logrus.NewEntry(logrus.StandardLogger()))
}
