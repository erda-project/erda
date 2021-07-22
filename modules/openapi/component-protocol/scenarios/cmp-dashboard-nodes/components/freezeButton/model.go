package freezeButton

import protocol "github.com/erda-project/erda/modules/openapi/component-protocol"

type FreezeButton struct {
	ctxBdl protocol.ContextBundle
	Type       string
	Props      Props
	Operations map[string]Operation `json:"operations,omitempty"`
}
type Props struct {
	Text    string
	Type    string
	Tooltip string
}
type Meta struct{
	NodeName string `json:"node_name"`
	ClusterName string `json:"cluster_name"`
}
type Operation struct {
	Key           string      `json:"key"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta"`
	Meta          interface{} `json:"meta"`
	ClickableKeys interface{} `json:"clickableKeys"`
	Command       Command     `json:"command,omitempty"`
}
type Command struct {
	Key    string `json:"key"`
	Target string `json:"target"`
}
