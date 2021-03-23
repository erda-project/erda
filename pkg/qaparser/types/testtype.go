package types

// 测试结果解析 parser
type TestParserType string

const (
	// 使用 maven-surefire-plugin 生成的 TEST-xxx.xml 格式进行解析
	Default TestParserType = "DEFAULT"
	// 使用 testng 插件 org.testng.reporters.XMLReporter 生成的格式进行解析, 需要在 pom.xml 中配置该 reporter
	NGTest TestParserType = "NGTEST"
	// 使用 junit 生成的 xml 格式进行解析
	JUnit TestParserType = "JUNIT"
)

func (t TestParserType) TPValue() string {
	return string(t)
}
