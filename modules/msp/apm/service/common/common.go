package common

import "github.com/erda-project/erda-proto-go/common/pb"

const (
	JavaProcessType   = "jvm_memory"
	NodeJsProcessType = "nodejs_memory"
)

var ProcessTypes = map[string]string{
	JavaProcessType:   pb.Language_java.String(),
	NodeJsProcessType: pb.Language_nodejs.String(),
}
