// +build dingding

package eventbox

// import (
// 	"bytes"
// 	"io/ioutil"
// 	"math/rand"
// 	"sync"
// 	"testing"
// 	"time"

// 	. "github.com/erda-project/erda/modules/eventbox/testutil"
// 	"github.com/erda-project/erda/pkg/httpclient"

// 	"github.com/stretchr/testify/assert"
// )

// func TestNormalDDOutput(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()
// 	content := GenContent("test1")
// 	resp, _, err := OutputDD(content)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 200, resp.StatusCode())
// }

// func TestTooManyNormalDDOutput(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()

// 	type respAndbody struct {
// 		resp *httpclient.Response
// 		body *bytes.Buffer
// 	}

// 	var wg sync.WaitGroup
// 	count := 30

// 	bufs := make(chan respAndbody, count)
// 	wg.Add(count)
// 	for i := 0; i < count; i++ {
// 		go func() {
// 			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
// 			resp, buf, err := OutputDD(GenContent("test1"))
// 			assert.Nil(t, err)
// 			bufs <- respAndbody{resp, buf}
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// 	close(bufs)

// 	for respBody := range bufs {
// 		all, err := ioutil.ReadAll(respBody.body)
// 		assert.Nil(t, err)
// 		if !respBody.resp.IsOK() {
// 			assert.Contains(t, string(all), "send too fast")
// 		}
// 	}
// }
