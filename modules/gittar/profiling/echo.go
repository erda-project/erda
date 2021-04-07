// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package profiling

import (
	"net/http/pprof"
	"runtime"

	"github.com/labstack/echo"
)

func Wrap(e *echo.Echo) {
	WrapGroup(e.Group(""))
}

// WrapGroup adds several routes from package `net/http/pprof` to *echo.Group object.
func WrapGroup(g *echo.Group) {
	routers := []struct {
		Method  string
		Path    string
		Handler echo.HandlerFunc
	}{
		{"GET", "/pprof/", IndexHandler()},
		{"GET", "/pprof/heap", HeapHandler()},
		{"GET", "/pprof/goroutine", GoroutineHandler()},
		{"GET", "/pprof/block", BlockHandler()},
		{"GET", "/pprof/allocs", AllocsHandler()},
		{"GET", "/pprof/threadcreate", ThreadCreateHandler()},
		{"GET", "/pprof/cmdline", CmdlineHandler()},
		{"GET", "/pprof/profile", ProfileHandler()},
		{"GET", "/pprof/symbol", SymbolHandler()},
		{"POST", "/pprof/symbol", SymbolHandler()},
		{"GET", "/pprof/trace", TraceHandler()},
		{"GET", "/pprof/mutex", MutexHandler()},
		{"GET", "/gc", GCHandler()},
	}

	for _, r := range routers {
		switch r.Method {
		case "GET":
			g.GET(r.Path, r.Handler)
		case "POST":
			g.POST(r.Path, r.Handler)
		}
	}
}

func GCHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		runtime.GC()
		ctx.String(200, "ok")
		return nil
	}
}

// IndexHandler will pass the call from /debug/pprof to pprof.
func IndexHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Index(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// HeapHandler will pass the call from /debug/pprof/heap to pprof.
func HeapHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("heap").ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

// GoroutineHandler will pass the call from /debug/pprof/goroutine to pprof.
func GoroutineHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("goroutine").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// BlockHandler will pass the call from /debug/pprof/block to pprof.
func BlockHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("block").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}

} // BlockHandler will pass the call from /debug/pprof/block to pprof.
func AllocsHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("allocs").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// ThreadCreateHandler will pass the call from /debug/pprof/threadcreate to pprof.
func ThreadCreateHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("threadcreate").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// MutexHandler will pass the call from /debug/pprof/mutex to pprof.
func MutexHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Handler("mutex").ServeHTTP(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// CmdlineHandler will pass the call from /debug/pprof/cmdline to pprof.
func CmdlineHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Cmdline(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// 持续30s，并生成一个文件供下载
func ProfileHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Profile(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// SymbolHandler will pass the call from /debug/pprof/symbol to pprof.
func SymbolHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Symbol(ctx.Response().Writer, ctx.Request())
		return nil
	}
}

// TraceHandler will pass the call from /debug/pprof/trace to pprof.
func TraceHandler() echo.HandlerFunc {
	return func(ctx echo.Context) error {
		pprof.Trace(ctx.Response().Writer, ctx.Request())
		return nil
	}
}
