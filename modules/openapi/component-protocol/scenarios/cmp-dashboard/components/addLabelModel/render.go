package addLabelModel

import (
	"context"
	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

func (a *AddLabelModel) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	return a.SetComponentValue(c)
}

// SetComponentValue transfer CpuInfoTable struct to Component
func (a *AddLabelModel) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
	)
	if state, err = common.ConvertToMap(a.State); err != nil {
		return err
	}
	c.State = state
	c.Props = getProps()
	return nil
}

func getProps() map[string][]Fields {
	return map[string][]Fields{
		"fields": {
			{
				Component: "select",
				Key:       "labelGroup",
				Label:     "分组",
				Required:  true,
				ComponentProps: map[string][]Options{
					"Options": {{
						Name:  "",
						Value: "",
					}, {
						Name:  "",
						Value: "",
					},
					},
				},
			}, {
				Component: "input",
				Key:       "name",
				Label:     "标签",
				Required:  true,
				Rules: []map[string]string{
					{"msg": "格式：ss", "pattern": "/^[.a-z\\u4e00-\\u9fa5A-Z0-9_-\\s]*$/"},
				},
			},
		},
	}
}
func getOperations() map[string]interface{} {
	return nil
}
func getState() State {
	return State{
		Visible:  false,
		FormData: nil,
	}
}
func RenderCreator() protocol.CompRender {
	return &AddLabelModel{
		Type:      "FormModal",
		Props:     getProps(),
		State:     getState(),
		Operation: getOperations(),
	}
}
