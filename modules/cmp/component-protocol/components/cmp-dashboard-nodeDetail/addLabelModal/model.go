package AddLabelModal

import (
	"context"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type AddLabelModal struct {
	*cptype.SDK
	Ctx    context.Context
	CtxBdl *bundle.Bundle
	base.DefaultProvider
	Type       string                `json:"type"`
	Props      Props                 `json:"props"`
	State      State                 `json:"state"`
	Operations map[string]Operations `json:"operations"`
}

type Props struct {
	Fields []Fields `json:"fields"`
	Title  string   `json:"title"`
}

type State struct {
	FormData map[string]string `json:"formData"`
	Visible  bool              `json:"visible"`
}

type Operations struct {
	Key    string `json:"key"`
	Reload bool   `json:"reload"`
}

type Fields struct {
	Key            string         `json:"key"`
	ComponentProps ComponentProps `json:"componentProps"`
	Label          string         `json:"label"`
	Component      string         `json:"component"`
	Rules          Rules          `json:"rules"`
	Required       bool           `json:"required"`
	RemoveWhen     [][]RemoveWhen
}

type RemoveWhen struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type ComponentProps struct {
	Options []Option `json:"options"`
}

type Option struct {
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	Children []Option `json:"children"`
}

type Rules struct {
	Msg     string `json:"msg"`
	Pattern string `json:"pattern"`
}
