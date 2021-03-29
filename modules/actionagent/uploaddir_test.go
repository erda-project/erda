package actionagent

//import (
//	"fmt"
//	"os"
//	"testing"
//)
//
//func TestAgent_UploadDir(t *testing.T) {
//	os.Setenv("DICE_OPENAPI_TOKEN", "fake_token")
//	agent := Agent{
//		EasyUse: EasyUse{
//			ContainerUploadDir: "/tmp/uploaddir",
//			OpenAPIAddr:        "openapi.default.svc.cluster.local:9529",
//		},
//	}
//	agent.uploadDir()
//	for _, err := range agent.Errs {
//		fmt.Println(err)
//	}
//}
