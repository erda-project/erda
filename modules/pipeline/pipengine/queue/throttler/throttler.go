package throttler

import (
	"sync"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/enhancedqueue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/snapshot"
)

type Throttler interface {
	// Name 节流阀的名字
	Name() string

	// AddQueue 幂等创建队列
	AddQueue(name string, window int64)

	// AddKeyToQueues 幂等将 key 同时插入指定队列；若队列不存在，则会首先幂等创建队列；若不同时插入，则可能直接被调度了
	AddKeyToQueues(key string, reqs []AddKeyToQueueRequest)

	PopPending(key string) (bool, []PopDetail)
	PopProcessing(key string) (bool, []PopDetail)

	snapshot.Snapshot
}

type throttler struct {
	name             string
	queueByName      map[string]*enhancedqueue.EnhancedQueue            // 该节流阀关心的所有增强队列
	keyRelatedQueues map[string]map[string]*enhancedqueue.EnhancedQueue // 所有 key 和队列的关联关系
	lock             sync.Mutex
}

func NewNamedThrottler(name string, initQueues map[string]int64) Throttler {
	t := throttler{
		name:             name,
		queueByName:      make(map[string]*enhancedqueue.EnhancedQueue),
		keyRelatedQueues: make(map[string]map[string]*enhancedqueue.EnhancedQueue),
		lock:             sync.Mutex{},
	}
	for name, window := range initQueues {
		t.AddQueue(name, window)
	}
	return &t
}

func (t *throttler) Name() string {
	return t.name
}

func (t *throttler) AddQueue(name string, window int64) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.addQueue(name, window)
}

func (t *throttler) AddKeyToQueues(key string, reqs []AddKeyToQueueRequest) {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, req := range reqs {
		t.addKeyToQueue(key, req)
	}
}

// PopPending 将指定 key 从所有关联队列 pending 弹出到 processing
// 若 key 不存在，返回可弹出
func (t *throttler) PopPending(key string) (bool, []PopDetail) {
	t.lock.Lock()
	defer t.lock.Unlock()

	var popDetails []PopDetail

	relatedQueues, ok := t.keyRelatedQueues[key]
	if !ok {
		// key 没有关联的增强队列，返回成功
		popDetails = append(popDetails, PopDetail{
			CanPop: true,
			Reason: "no related queues",
		})
		return true, popDetails
	}

	canPop := true
	for qName, eq := range relatedQueues {
		poppedKey := eq.PopPending(true)
		if poppedKey != key {
			popDetails = append(popDetails, newPopDetail(qName, false, "cannot pop pending now, waiting for next time"))
			canPop = false
			continue
		}
		popDetails = append(popDetails, newPopDetail(qName, true, ""))
	}

	// 可弹出
	if canPop {
		// 遍历弹出
		for _, eq := range relatedQueues {
			eq.PopPending()
		}
	}
	return canPop, popDetails
}

// PopProcessing
func (t *throttler) PopProcessing(key string) (bool, []PopDetail) {
	t.lock.Lock()
	defer t.lock.Unlock()

	relatedQueues, ok := t.keyRelatedQueues[key]
	if !ok {
		return true, nil
	}

	var popDetails []PopDetail
	canPop := true
	for qName, eq := range relatedQueues {
		poppedKey := eq.PopProcessing(key, true)
		if poppedKey != key {
			popDetails = append(popDetails, newPopDetail(qName, false, "cannot pop processing now, waiting for next time"))
			canPop = false
			continue
		}
		popDetails = append(popDetails, newPopDetail(qName, true, ""))
	}

	// 可弹出
	if canPop {
		// 遍历弹出
		for _, eq := range relatedQueues {
			eq.PopProcessing(key)
		}
		// cleanup
		delete(t.keyRelatedQueues, key)
	}
	return canPop, popDetails
}
