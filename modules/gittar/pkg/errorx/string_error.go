package errorx

import "fmt"

// StringError 简单的字符串型错误
type StringError string

// New 创建一个具有文本信息的错误
func New(text string) error {
	return StringError(text)
}

// Errorf 创建一个具有文本信息的错误
func Errorf(format string, args ...interface{}) error {
	return StringError(fmt.Sprintf(format, args...))
}

func (err StringError) Error() string {
	return string(err)
}
