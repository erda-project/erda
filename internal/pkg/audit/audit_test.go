package audit

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type ctxKey string

func TestGetContextEntryMapEmpty(t *testing.T) {
	t.Run("no_context_data", func(t *testing.T) {
		ctx := context.Background()
		got := GetContextEntryMap(ctx)
		if got != nil {
			t.Fatalf("expected nil, got: %#v", got)
		}
	})

	t.Run("empty_opts", func(t *testing.T) {
		ctx := withOptionDataContext(context.Background(), &optionContextData{})
		got := GetContextEntryMap(ctx)
		if got != nil {
			t.Fatalf("expected nil, got: %#v", got)
		}
	})
}

func TestGetContextEntryMapEntriesAndFetchers(t *testing.T) {
	base := context.WithValue(context.Background(), ctxKey("k"), "ctxv")
	ctx := withOptionDataContext(base, &optionContextData{})

	ContextEntry(ctx, "plain", "v")
	ContextEntry(ctx, "fetcher", ValueFetcher(func() interface{} { return 123 }))
	ContextEntry(ctx, "ctx_fetcher", ValueFetcherWithContext(func(ctx context.Context) (interface{}, error) {
		return ctx.Value(ctxKey("k")), nil
	}))
	ContextEntry(ctx, "func", func() interface{} { return "f" })
	ContextEntry(ctx, "func_ctx", func(ctx context.Context) (interface{}, error) { return "g", nil })
	ContextEntry(ctx, "err", ValueFetcherWithContext(func(ctx context.Context) (interface{}, error) {
		return nil, errors.New("boom")
	}))
	ContextEntryMap(ctx, map[string]interface{}{"mapKey": "mapVal"})

	got := GetContextEntryMap(ctx)
	if got == nil {
		t.Fatal("expected non-nil map")
	}

	want := map[string]interface{}{
		"plain":       "v",
		"fetcher":     123,
		"ctx_fetcher": "ctxv",
		"func":        "f",
		"func_ctx":    "g",
		"mapKey":      "mapVal",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected map.\nwant: %#v\n got: %#v", want, got)
	}
}
