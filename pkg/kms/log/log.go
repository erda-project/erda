package log

import (
	"context"

	"github.com/sirupsen/logrus"
)

const diceTraceID = "dice-trace-id"

func WithTraceID(ctx context.Context) *logrus.Entry {
	return logrus.WithField(diceTraceID, ctx.Value(diceTraceID))
}
