package prechecktype

import "context"

const (
	CtxResultKeyCrossCluster = "cross_cluster" // bool
)

const (
	ctxResultKey = "result"
)

func InitContext() context.Context {
	return context.WithValue(context.Background(), ctxResultKey, map[string]interface{}{})
}

func PutContextResult(ctx context.Context, k string, v interface{}) {
	ctx.Value(ctxResultKey).(map[string]interface{})[k] = v
}

func GetContextResult(ctx context.Context, k string) interface{} {
	return ctx.Value(ctxResultKey).(map[string]interface{})[k]
}
