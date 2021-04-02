package subscriber

import (
	"github.com/erda-project/erda/modules/eventbox/types"
)

type Subscriber interface {
	// 各个实现自己解析 dest
	// 返回 []error , 是因为发送消息的目的可能是多个
	// dest: marshaled string
	// content: marshaled string
	Publish(dest string, content string, time int64, m *types.Message) []error
	Status() interface{}
	Name() string
}
