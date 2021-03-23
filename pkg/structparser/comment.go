package structparser

import (
	"reflect"
)

// getComment 获取 `t` 的 `fieldname` 对应的注释(comment)
// `t` 是 struct value
func getComment(t interface{}, fieldname string) string {
	tp := reflect.TypeOf(t)
	value := reflect.ValueOf(t)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	method := value.MethodByName("Desc_" + tp.Name())

	if (method == reflect.Value{}) {
		return ""
	}
	return method.Call([]reflect.Value{reflect.ValueOf(fieldname)})[0].String()
}
