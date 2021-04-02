/*
对外提供 API 用于注册特定 label 对 其他 label 的关系
暂定这个特定 label 的 key 为 "REGISTERED_LABEL".(constant.RegisterLabelKey)

e.g.
PUT /dice/eventbox/register/<VALUE>
body: {"labels": map[string]string{"<label1>":"<value1>", "<label2>":"<value2>"}}

之后在发送消息的时候，带上 label : {"REGISTERED_LABEL":"<VALUE>"},
相当于 带上了 上面所注册的所有 labels

*/
package register
