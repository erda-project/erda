package snippetsvc

import (
	"encoding/json"
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func Test_ActionJson(t *testing.T) {

	var str = "{\"alias\":\"account-login\",\"type\":\"api-test\",\"description\":\"执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\",\"version\":\"2.0\",\"params\":{\"asserts\":[{\"arg\":\"status\",\"operator\":\"=\",\"value\":\"200\"},{\"arg\":\"sessionId\",\"operator\":\"not_empty\"}],\"headers\":[{\"key\":\"Content-Type\",\"value\":\"application/x-www-form-urlencoded\"}],\"method\":\"POST\",\"name\":\"登录\",\"out_params\":[{\"expression\":\".sessionid\",\"key\":\"sessionId\",\"source\":\"body:json\"},{\"key\":\"status\",\"source\":\"status\"},{\"expression\":\".id\",\"key\":\"userId\",\"source\":\"body:json\"}],\"params\":[{\"key\":\"username\",\"value\":\"${params.username}\"},{\"key\":\"password\",\"value\":\"${params.password}\"}],\"url\":\"${params.openapi_addr}/login\"},\"resources\":{},\"displayName\":\"接口测试\",\"logoUrl\":\"//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\"}"

	var action apistructs.PipelineYmlAction
	err := json.Unmarshal([]byte(str), &action)
	if err != nil {
		t.Fail()
	}

	params := action.Params
	if params == nil {
		t.Fail()
	}

	outParamsBytes, err := json.Marshal(action.Params["out_params"])
	if err != nil {
		t.Fail()
	}

	var outParams []apistructs.APIOutParam
	err = json.Unmarshal(outParamsBytes, &outParams)
	if err != nil {
		t.Fail()
	}

}
