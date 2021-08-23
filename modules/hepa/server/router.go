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

package server

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/common"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
	reset     = string([]byte{27, 91, 48, 109})
)

var once sync.Once
var router *gin.Engine

type Controller func(*gin.Context, []byte) (int, []byte)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

func timeFormat(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				stack := stack(3)
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					log.Errorf("%s\n%s%s", err, string(httprequest), reset)
				} else {
					log.Panicf("[Recovery] %s panic recovered:\n%s\n%s\n%s%s",
						timeFormat(time.Now()), string(httprequest), err, stack, reset)
				}

				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					_ = c.Error(err.(error))
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}

}

func CreateSingleton(logger *log.Logger) *gin.Engine {
	once.Do(func() {
		router = gin.New()
		router.Use(common.AccessWrap(logger), Recovery())
	})
	return router
}

func GetSingleton() *gin.Engine {
	if router == nil {
		panic("router not init")
	}
	return router
}

func Start(server *http.Server) error {
	if router == nil {
		return errors.New("router not init")
	}
	server.Handler = router
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func BindRawApi(location, method string, controller Controller, asyncController ...Controller) {
	bindApi("", location, method, controller, asyncController...)
}

func BindApi(location, method string, controller Controller, asyncController ...Controller) {
	bindApi(API_GATEWAY_PREFIX, location, method, controller, asyncController...)
}

func BindOpenApi(location, method string, controller Controller, asyncController ...Controller) {
	bindApi(OPENAPI_PREFIX, location, method, controller, asyncController...)
}

func doRecover() {
	if r := recover(); r != nil {
		log.Errorf("recovered from: %+v ", r)
		debug.PrintStack()
	}
}

func bindApi(prefix, location, method string, controller Controller, asyncController ...Controller) {
	if router == nil {
		panic("router not init")
	}
	router.Handle(method, prefix+location, func(c *gin.Context) {
		c.Set("startTime", time.Now().Unix())
		reqBody, err := c.GetRawData()
		if err != nil {
			log.Error(err)
			c.String(http.StatusInternalServerError, "")
			return
		}
		//project auth check
		projectId := c.Query("projectId")
		if projectId != "" {
			succ, err := CheckAuth(c, projectId)
			if err != nil {
				log.Error(err)
				c.String(http.StatusForbidden, "")
				return
			}
			if !succ {
				c.String(http.StatusForbidden, "")
				return
			}
		}
		status, respBody := controller(c, reqBody)
		c.Data(status, "application/json; charset=utf-8", respBody)
		c.Set("reqBody", string(reqBody))
		c.Set("respBody", string(respBody))
		if len(asyncController) > 0 && status == http.StatusOK {
			doAsyncI, ok := c.Get("do_async")
			if !ok {
				log.Info("no need do async")
				return
			}
			doAsync, ok := doAsyncI.(bool)
			if !ok {
				log.Error("acquire doAsync failed")
				return
			}
			if !doAsync {
				log.Info("no need do async")
				return
			}
			copyContext := c.Copy()
			go func() {
				defer doRecover()
				asyncController[0](copyContext, reqBody)
			}()
		}
	})
}
