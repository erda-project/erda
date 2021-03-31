// TODO:
// INPUT test:
// - 写入错误格式的 message 到 etcd
// - 同时起 2 个 eventbox， 写入错误格式的 message 到 etcd
// - 设置 dispatcher 的 goroutinepool 最大为 5 ， 使用延时 1s 的 backend 来分发数据，同时写入 etcd 正常的 message 大于 5
// - 发送错误格式 message 到 httpserver
// - 同时起 2 个 eventbox， 写入正常格式的 message 到 httpserver
// - 同时起 2 个 eventbox， 写入错误格式的 message 到 httpserver
// - 设置 dispatcher 的 goroutinepool 最大为 5 ， 使用延时 1s 的 backend 来分发数据，同时写入 httpserver 正常的 message 大于 5

// output test:

// DONE:
// INPUT test:
// - 写入正常格式的 message 到 etcd
// - 启动后， 处理 etcd 历史消息
// - 发送正常格式 message 到 httpserver
// - 同时起 2 个 eventbox， 写入正常格式的 message 到 etcd
// - 同时发送 100 个 正常格式 message 到 etcd
// - 同时发送 100 个 正常格式 message 到 httpserver

// output test:
// - 写入正常格式的 message 到 dingding
// - 写入过多的正常格式的 message 到 dingding

package eventbox

// import (
// 	"bytes"
// 	"io/ioutil"
// 	"net/url"
// 	"os"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/erda-project/erda/modules/eventbox/conf"
// 	"github.com/erda-project/erda/modules/eventbox/dispatcher"
// 	"github.com/erda-project/erda/modules/eventbox/register"
// 	. "github.com/erda-project/erda/modules/eventbox/testutil"
// 	"github.com/erda-project/erda/modules/eventbox/types"
// 	"github.com/erda-project/erda/pkg/httpclient"

// 	"github.com/sirupsen/logrus"
// 	"github.com/stretchr/testify/assert"
// )

// var originListenAddr string

// func init() {
// 	originListenAddr = conf.ListenAddr()
// }

// var dp, _ = dispatcher.New()
// var dp2, _ = dispatcher.New()
// var dpDefault, _ = dispatcher.New()

// func startDefault() {
// 	startEventbox(originListenAddr, dpDefault)
// }

// func stopDefault() {
// 	stopEventbox(dpDefault)
// }

// func startEventbox(listenAddr string, dp dispatcher.Dispatcher) {
// 	os.Setenv("LISTEN_ADDR", listenAddr)
// 	conf.C.ListenAddr = listenAddr
// 	go func() {
// 		dp.Start()
// 	}()
// 	time.Sleep(1 * time.Second)

// 	logrus.Info("start eventbox")
// }

// func stopEventbox(dp dispatcher.Dispatcher) {
// 	dp.Stop()
// 	logrus.Info("stop eventbox")
// }

// func TestNormalEtcdInput(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()

// 	content := GenContent("test")
// 	err := InputEtcd(content)
// 	assert.Nil(t, err)
// 	time.Sleep(2 * time.Second)
// 	assert.True(t, FakeFileContains(content))
// }

// func TestRestore(t *testing.T) {
// 	defer stopEventbox(dp)
// 	assert.Nil(t, InputEtcd(GenContent("test1")))
// 	assert.Nil(t, InputEtcd(GenContent("test2")))
// 	assert.Nil(t, InputEtcd(GenContent("test3")))
// 	assert.Nil(t, InputEtcd(GenContent("test4")))
// 	assert.Nil(t, InputEtcd(GenContent("test5")))
// 	startEventbox(":23334", dp)
// 	time.Sleep(2 * time.Second)
// 	assert.True(t, IsCleanEtcd())

// }

// func Test2EventboxWithNormalEtcdInput(t *testing.T) {
// 	startEventbox(":23335", dp)
// 	time.Sleep(1 * time.Second) // make sure dp got lock
// 	startEventbox(":23336", dp2)
// 	stopEventbox(dp2)

// 	content := GenContent("test1")
// 	assert.Nil(t, InputEtcd(content))
// 	time.Sleep(1 * time.Second)
// 	assert.True(t, IsCleanEtcd()) // has been consumed
// 	assert.True(t, FakeFileContains(content))

// 	// dp3, _ := dispatcher.New()
// 	startEventbox(":23337", dp2)
// 	stopEventbox(dp)

// 	content = GenContent("test1")
// 	assert.Nil(t, InputEtcd(content))
// 	time.Sleep(1 * time.Second)
// 	assert.True(t, IsCleanEtcd())
// 	assert.True(t, FakeFileContains(content))
// 	stopEventbox(dp2)
// }

// func TestNormalHTTPInput(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()
// 	content := GenContent("test1")
// 	resp, buf, err := InputHTTP(content)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 200, resp.StatusCode())
// 	if resp.StatusCode() != 200 {
// 		all, _ := ioutil.ReadAll(buf)
// 		logrus.Errorf(string(all))
// 	}
// 	assert.True(t, FakeFileContains(content))
// }

// // func Test100MessageToEtcd(t *testing.T) {
// // 	startDefault()
// // 	defer stopDefault()
// // 	var wg sync.WaitGroup
// // 	wg.Add(100)
// // 	for i := 0; i < 100; i++ {
// // 		go func() {
// // 			content := GenContent("test1")
// // 			assert.Nil(t, InputEtcd(content))
// // 			wg.Done()
// // 		}()
// // 	}
// // 	wg.Wait()
// // 	time.Sleep(1 * time.Second) // 假设 1s 内处理完
// // 	assert.True(t, IsCleanEtcd())
// // }

// func Test100MessageToHTTP(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()
// 	var wg sync.WaitGroup
// 	wg.Add(100)
// 	for i := 0; i < 100; i++ {
// 		go func() {
// 			content := GenContent("test1")
// 			resp, buf, err := InputHTTP(content)
// 			assert.Nil(t, err)
// 			if err != nil {
// 				logrus.Errorf("%v\n", err.(*url.Error).Err)

// 			}
// 			if err == nil {
// 				assert.True(t, resp.IsOK())
// 				if !resp.IsOK() {
// 					all, _ := ioutil.ReadAll(buf)
// 					logrus.Error(string(all))
// 				}
// 			}
// 			wg.Done()
// 		}()
// 	}
// 	wg.Wait()
// 	assert.True(t, IsCleanEtcd())
// }

// func TestMessageWithRegisteredLabel(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()
// 	content := GenContent("TestMessageWithRegisteredLabel")
// 	resp, _, err := InputHTTPRaw(content, conf.ListenAddr()[1:], map[types.LabelKey]interface{}{"REGISTERED_LABEL": []string{"TEST_LABEL"}})
// 	assert.Nil(t, err)
// 	assert.Equal(t, 200, resp.StatusCode())
// 	assert.True(t, FakeFileContains(content))
// }

// func TestMessageWithNewRegisterLabel(t *testing.T) {
// 	startDefault()
// 	defer stopDefault()
// 	resp, err := httpclient.New().Put("127.0.0.1:9528").Path("/api/dice/eventbox/register").Header("Accept", "application/json").JSONBody(register.PutRequest{"TEST_REGISTER_LABEL/LVL1/LVL2", map[types.LabelKey]interface{}{"FAKE": ""}}).Do().DiscardBody()
// 	assert.Nil(t, err)
// 	if resp != nil {
// 		assert.True(t, resp.IsOK())
// 	}
// 	content := GenContent("TestMessageWithNewRegisterLabel")
// 	resp, _, err = InputHTTPRaw(content, conf.ListenAddr()[1:], map[types.LabelKey]interface{}{"REGISTERED_LABEL": []string{"TEST_REGISTER_LABEL/LVL1/LVL2"}})
// 	assert.Nil(t, err)
// 	assert.Equal(t, 200, resp.StatusCode())
// 	assert.True(t, FakeFileContains(content))

// 	var buf bytes.Buffer
// 	resp, err = httpclient.New().Delete("127.0.0.1:9528").Path("/api/dice/eventbox/register").Header("Accept", "application/json").JSONBody(
// 		register.DelRequest{"TEST_REGISTER_LABEL/LVL1/LVL2"}).Do().Body(&buf)
// 	assert.Nil(t, err)

// 	assert.Equal(t, 200, resp.StatusCode())

// }
