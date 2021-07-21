package tab

const (
	CPU_TAB = "cpu"
	CPU_TAB_ZH = "cpu分析"

	MEM_TAB = "mem"
	MEM_TAB_ZH = "mem分析"

)
type SteveTab struct {
	Type       string     `json:"type"`
	Props      Props      `json:"props,omitempty"`
	State      PropsState `json:"state,omitempty"`
	Operations map[string]interface{}
}
type Props struct {
	TabMenu []MenuPair `json:"tab_menu,omitempty"`
}
type MenuPair struct {
	key string
	name string
}
type PropsState struct {
	ActiveKey string `json:"active_key"`
}