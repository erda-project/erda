package apistructs

type TestPlanActiveKey string

const (
	ConfigTestPlanActiveKey  TestPlanActiveKey = "Config"
	ExecuteTestPlanActiveKey TestPlanActiveKey = "Execute"
)

func (s TestPlanActiveKey) String() string {
	return string(s)
}
