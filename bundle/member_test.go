package bundle

//import (
//	"os"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/erda-project/erda/apistructs"
//)
//
//func TestGetMemberByToken(t *testing.T) {
//	// set env
//	os.Setenv("CMDB_ADDR", "http://cmdb.default.svc.cluster.local:9093/")
//
//	bdl := New(WithCMDB())
//	request := &apistructs.GetMemberByTokenRequest{
//		Token: "333333",
//	}
//
//	member, err := bdl.GetMemberByToken(request)
//	assert.Nil(t, err)
//	println(member.Email)
//}
