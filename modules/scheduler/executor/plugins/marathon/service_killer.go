package marathon

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/events"
	_ "github.com/erda-project/erda/pkg/monitor"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	PERIOD     = 60
	TIMES      = 3
	PROTECTION = 600
)

type judgement struct {
	record map[string][]int64
	sync.RWMutex
}

// 监测一段时间内多次被杀死的实例, 实例所隶属的 service 会被送到这里处死
func (m *Marathon) SuspendApp(ch chan string) {
	// record 的 key 实例所在 service, value 是实例被杀时间
	j := &judgement{record: make(map[string][]int64)}

	go compactRecord(j)

	for appID := range ch {
		// patch /v2/apps/{app_id} to scale the instances to zero
		var ag AppGet
		var b bytes.Buffer
		logrus.Infof("got appID(%s) to judge", appID)

		srvInstances := 1
		resp, err := m.client.Get(m.addr).
			Path("/v2/apps/" + appID).
			Do().
			Body(&b)
		if err != nil {
			logrus.Errorf("marathon get appID(%s) failed, err: %v", appID, err)
			continue
		}
		if !resp.IsOK() {
			logrus.Errorf("marathon get appID(%s) status code: %v, body: %v", appID, resp.StatusCode(), b.String())
			continue
		}
		r := bytes.NewReader(b.Bytes())
		if err := json.NewDecoder(r).Decode(&ag); err != nil {
			return
		}
		if ag.App.Instances == 0 {
			continue
		}
		srvInstances = ag.App.Instances
		logrus.Infof("got appID(%s) instances number: %v", appID, srvInstances)

		// 记录, 有罪则审判
		if !determineGuilty(appID, srvInstances, j) {
			continue
		}

		// 对确认有罪但在保护期内(10分钟)的服务豁免, 释放
		if inProtection(appID) {
			continue
		}

		// suspend the service, scale the instance(replica) to 0
		sa := ShortApp{Instances: 0}
		resp, err = m.client.Patch(m.addr).
			Path("/v2/apps/" + appID).
			JSONBody(sa).
			Do().
			Body(&b)
		if err != nil {
			logrus.Errorf("marathon patch appID(%s) failed, err: %v", appID, err)
			continue
		}
		if !resp.IsOK() {
			logrus.Errorf("marathon patch appID(%s) status code: %v, body: %v", appID, resp.StatusCode(), b.String())
			continue
		}

		// "appId":"/runtimes/v1/services/staging-2725/pmp-backend"
		logrus.Errorf("[alert] Sentence service(appID: %s, instances: %v) to death, killed time: %v",
			appID, srvInstances, j.record[appID])
	}
}

func generateEtcdKey(appID string) (string, error) {
	strs := strings.Split(appID, "/")
	if len(strs) < 5 {
		return "", errors.Errorf("appID(%s) format is wrong", appID)
	}
	// strs[3]: services
	// strs[4]: staging-2725
	return "/dice/service/" + strs[3] + "/" + strs[4], nil
}

// 保护 PROTECTION 秒内被修改过的 runtime 下的所有 service
func inProtection(appID string) bool {
	key, err := generateEtcdKey(appID)
	if err != nil {
		logrus.Errorf(err.Error())
		return true
	}
	runtime := apistructs.ServiceGroup{}
	em := events.GetEventManager()
	if err = em.MemEtcdStore.Get(context.Background(), key, &runtime); err != nil {
		logrus.Errorf("get runtime(%s) from memetcd error: %v", key, err)
		return true
	}

	if time.Now().Unix()-runtime.LastModifiedTime < PROTECTION {
		logrus.Infof("appID(%s) is in protection, modified time: %v", appID, runtime.LastModifiedTime)
		return true
	}
	return false
}

// 裁定是否有罪，有罪定义为 $(PERIOD) 秒内有 $(TIMES) 次被杀
func determineGuilty(appID string, instance int, j *judgement) bool {
	t := time.Now().Unix()
	j.RLock()
	array, ok := j.record[appID]
	j.RUnlock()
	if !ok {
		j.Lock()
		j.record[appID] = []int64{t}
		j.Unlock()
		return false
	}
	if len(array) >= instance*TIMES {
		permutation(array, t)
		j.Lock()
		j.record[appID] = array
		j.Unlock()

		begin := len(array) - instance*TIMES
		if t-array[begin] <= PERIOD && t-array[begin] >= 0 {
			return true
		}
	} else {
		array = append(array, t)
		j.Lock()
		j.record[appID] = array
		j.Unlock()
	}
	return false
}

func permutation(s []int64, v int64) []int64 {
	for i := 1; i < len(s); i++ {
		s[i-1] = s[i]
	}
	s[len(s)-1] = v
	return s
}

// 删除过期的犯罪记录
func compactRecord(j *judgement) {
	for {
		select {
		case <-time.After(1 * time.Hour):
			logrus.Infof("going to clear criminal record, len: %v", len(j.record))
			for k, v := range j.record {
				if l := len(v); l > 0 && time.Now().Unix()-v[l-1] > 3600 {
					j.Lock()
					delete(j.record, k)
					j.Unlock()
					logrus.Infof("delete service(%s)'s criminal records", k)
				}
			}
		}
	}
}
