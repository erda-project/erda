package errorx

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

// 最大可纪录的调用堆栈的深度
var MaxStackDepth = 5

type TracedError interface {
	error
	Callers() []uintptr // 返回堆栈调用信息
	Unpack() error      // 返回去掉堆栈调用信息的错误
}

// TracedError 记录调用堆栈信息的错误类型
type tracedError struct {
	err   interface{}
	stack []uintptr
}

// NewTracedError 创建一个具有堆栈调用信息的错误
func NewTracedError(err interface{}) TracedError {
	if err == nil {
		return nil
	}
	stack := make([]uintptr, MaxStackDepth)
	stack = stack[:runtime.Callers(2, stack[:])]
	return &tracedError{
		err:   err,
		stack: stack,
	}
}

func (err *tracedError) Error() string {
	buf := bytes.Buffer{}
	switch val := err.err.(type) {
	case string:
		buf.WriteString(val)
		buf.WriteString(" :\n")
	case error:
		buf.WriteString(val.Error())
		buf.WriteString(" :\n")
	case nil:
		return ""
	}
	for _, pc := range err.stack {
		if pc == 0 {
			continue
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, lineNum := fn.FileLine(pc - 1)
		if lastslash := strings.LastIndex(file, "/"); lastslash >= 0 {
			file = file[lastslash+1:]
		}
		name := fn.Name()
		if lastslash := strings.LastIndex(name, "/"); lastslash >= 0 {
			name = name[lastslash+1:]
		}
		buf.WriteString(fmt.Sprintf("* [%s:%d]	%s\n", file, lineNum, name))
	}
	return string(buf.Bytes()[:buf.Len()-1])
}

func (err *tracedError) Callers() []uintptr {
	return err.stack
}

func (err *tracedError) Unpack() error {
	switch val := err.err.(type) {
	case string:
		return StringError(val)
	case error:
		return val
	case nil:
		return nil
	}
	return err
}
