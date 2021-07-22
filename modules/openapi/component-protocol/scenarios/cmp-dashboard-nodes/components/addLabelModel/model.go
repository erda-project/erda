package addLabelModel

type AddLabelModel struct {
	Type      string                 `json:"type"`
	Props     map[string][]Fields `json:"props"`
	State     State                  `json:"state"`
	Operation map[string]interface{} `json:"operation"`
}

type State struct {
	Visible  bool        `json:"visible"`
	FormData interface{} `json:"form_data"`
}

type Fields struct {
	Key            string
	Component      string
	Label          string
	Required       bool
	ComponentProps map[string][]Options
	Rules          []map[string]string
}
type Options struct {
	Name string `json:"name"`
	Value string `json:"value"`
}

type Rule struct {
	Msg     string `json:"msg"`
	Pattern string `json:"pattern"`
}
