// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"modernc.org/mathutil"
)

type (
	Values    []Value
	EntryType int
)

const (
	UndefinedType EntryType = iota
	IntType
	BoolType
	FloatType
	StringType
	UnsignedType
	ByteType
	ByteSliceType

	minCacheSize = 2
)

var (
	EvictError              = errors.New("cache memory evicted")
	NilPtrError             = errors.New("nil ptr error")
	ParamIllegalError       = errors.New("size of segments illegal")
	EntryTypeDifferentError = errors.New("entry and cacheValues type different")
	WrongEntryKeyError      = errors.New("entry key is wrong")
	InitEntryFailedError    = errors.New("entry initialized failed")
	ValueNotFoundError      = errors.New("cacheValues not found")
	KeyNotFoundError        = errors.New("key not found")
	KeyTooLongError         = errors.New("key too long")
	SizeTooSmallError       = errors.New("total size too small")
	ValueTypeNotFoundError  = errors.New("cacheValues type not found")
	IllegalCacheSize        = errors.New("illegal cache size")
)

// Cache implement concurrent safe cache with LRU and ttl strategy.
type Cache struct {
	store *store
	log   *logrus.Logger
	k     *keyBuilder
}

type segment struct {
	pairs   []*pair
	tmp     *pair
	length  int
	mapping map[string]int
	maxSize int64
	used    int64
	nextIdx int64
}

func lowBit(x int) int {
	return x & (-x)
}

func getSegLen(size int, log *logrus.Logger) int {
	if size^lowBit(size) != 0 {
		i := 1
		for 1<<i < size {
			i++
		}
		log.Warnf("%d is not an idempotent of 2. reset size to %d", size, 1<<i)
		return 1 << i
	}
	return size
}
func newPairs(maxSize int64, segNum int) []*segment {

	ps := make([]*segment, segNum)
	// suppose average value size is 16B.
	// max length of pairs can not be larger than 1024.
	// if each size of values all less than 16B ,such as bool value,
	// memory can not be totally used
	pairLen := maxSize >> 4
	for i := 0; i < segNum; i++ {
		p := make([]*pair, mathutil.Min(int(pairLen), 1024))
		for j := range p {
			p[j] = &pair{}
		}
		ps[i] = &segment{
			pairs:   p,
			tmp:     &pair{},
			length:  0,
			mapping: map[string]int{},
			maxSize: maxSize,
			used:    0,
		}
	}
	return ps
}

type pair struct {
	key              string
	value            Values
	entryType        EntryType
	overdueTimestamp int64
	idx              int64
}

// store implement LRU strategy
type store struct {
	segs   []*segment
	locks  []*sync.RWMutex
	log    *logrus.Logger
	key    []byte
	segNum int
}

func (seg *segment) Len() int {
	return seg.length
}

func (seg *segment) Less(i, j int) bool {
	return seg.pairs[i].idx > seg.pairs[j].idx
}

func (seg *segment) Swap(i, j int) {
	mj := seg.mapping[seg.pairs[j].key]
	mi := seg.mapping[seg.pairs[i].key]
	seg.mapping[seg.pairs[i].key] = mj
	seg.mapping[seg.pairs[j].key] = mi
	seg.pairs[i], seg.pairs[j] = seg.pairs[j], seg.pairs[i]
}

func (seg *segment) Push(x interface{}) {
	return
}

func (seg *segment) Pop() interface{} {
	return nil
}

func newLocks(num int) []*sync.RWMutex {
	ls := make([]*sync.RWMutex, num)
	for i := range ls {
		ls[i] = &sync.RWMutex{}
	}
	return ls
}

func newStore(segSize int64, segLen int, logger *logrus.Logger) *store {
	return &store{
		log:    logger,
		segs:   newPairs(segSize, segLen),
		locks:  newLocks(segLen),
		key:    make([]byte, 1024),
		segNum: segLen,
	}
}

func (s *store) write(id int) error {
	ps := s.segs[id]
	newPair := s.segs[id].tmp

	needSize, _ := newPair.getEntrySize()
	if ps.maxSize < needSize {
		s.log.Errorf("evict cache size,try next")
		return EvictError
	}
	usage := ps.used
	for ps.maxSize-usage < needSize {
		p := ps.pairs[0]
		s.remove(id, p.key)
		entrySize, _ := p.getEntrySize()
		usage -= entrySize
		//s.log.Warnf("memory not sufficient ,%v has poped", p.key)
	}
	usage += needSize
	idx := ps.Len()
	ps.pairs[idx].key = newPair.key
	ps.pairs[idx].value = newPair.value
	ps.pairs[idx].overdueTimestamp = newPair.overdueTimestamp
	ps.mapping[newPair.key] = idx
	heap.Push(ps, ps.pairs[idx])
	ps.length++
	ps.used = usage
	return nil
}

func (s *store) remove(id int, key string) (*pair, error) {
	var (
		idx int
		ok  bool
	)
	ps := s.segs[id]
	if idx, ok = ps.mapping[key]; !ok {
		return nil, ValueNotFoundError
	}
	heap.Remove(ps, idx)
	ps.length--
	p := ps.pairs[ps.Len()]
	cacheSize, _ := p.getEntrySize()
	delete(ps.mapping, p.key)
	ps.used -= cacheSize
	return p, nil
}

type CmpCache interface {
	Remove(key string) (Values, error)
	Set(key string, value Values, overdueTimeStamp int64) error
	IncreaseSize(size int64)
	DecrementSize(size int64) error
	Get(key string) (Values, bool, error)
}

// updatePair update pair
func (seg *segment) updatePair(key string, newValues Values, overdueTimestamp int64) error {
	length := len(newValues)
	if length == 0 {
		return InitEntryFailedError
	}
	seg.tmp.overdueTimestamp = overdueTimestamp
	seg.tmp.value = newValues
	seg.tmp.key = key
	seg.tmp.idx = seg.nextIdx
	seg.nextIdx++
	return nil
}

// getEntrySize returns total size of values.
func (p *pair) getEntrySize() (int64, error) {
	if p.value == nil {
		return 0, NilPtrError
	}
	var usage = int64(0)
	for _, v := range p.value {
		usage += v.Size()
	}
	return usage, nil
}

type Value interface {
	// String returns string
	String() string
	// Type returns type of value
	Type() EntryType
	// Size returns size of value
	Size() int64
	// Value returns any type
	Value() interface{}
}

type ByteValue struct {
	value byte
}

func (b ByteValue) String() string {
	return string(b.value)
}

func (b ByteValue) Type() EntryType {
	return ByteType
}

func (b ByteValue) Size() int64 {
	return 9
}

func (b ByteValue) Value() interface{} {
	return b.value
}

type ByteSliceValue struct {
	value []byte
}

func (b ByteSliceValue) String() string {
	return string(b.value)
}

func (b ByteSliceValue) Type() EntryType {
	return ByteSliceType
}

func (b ByteSliceValue) Size() int64 {
	return int64(len(b.value))
}

func (b ByteSliceValue) Value() interface{} {
	return b.value
}

type FloatValue struct {
	value float64
}

func (f FloatValue) Size() int64 {
	return 16
}

func (f FloatValue) String() string {
	return fmt.Sprintf("%f", f.value)
}

func (f FloatValue) Type() EntryType {
	return FloatType
}

func (f FloatValue) Value() interface{} {
	return f.value
}

type IntValue struct {
	value int64
}

func (i IntValue) Size() int64 {
	return 16
}

func (i IntValue) String() string {
	return fmt.Sprintf("%v", i.value)
}

func (i IntValue) Type() EntryType {
	return IntType
}

func (i IntValue) Value() interface{} {
	return i.value
}

type StringValue struct {
	value string
}

func (s StringValue) Size() int64 {
	return 8 + int64(len(s.value))
}

func (s StringValue) String() string {
	return s.value
}

func (s StringValue) Type() EntryType {
	return StringType
}

func (s StringValue) Value() interface{} {
	return s.value
}

type UnsignedValue struct {
	value uint64
}

func (u UnsignedValue) Size() int64 {
	return 16
}

func (u UnsignedValue) String() string {
	return fmt.Sprintf("%v", u.value)
}

func (u UnsignedValue) Type() EntryType {
	return UnsignedType
}

func (u UnsignedValue) Value() interface{} {
	return u.value
}

type BoolValue struct {
	value bool
}

func (b BoolValue) Size() int64 {
	return 9
}

func (b BoolValue) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b BoolValue) Type() EntryType {
	return BoolType
}

func (b BoolValue) Value() interface{} {
	return b.value
}

// New returns cache.
// parma size means memory cache can use.
func New(size, segSize int64) (*Cache, error) {
	log := logrus.New()
	segNum := int(size / segSize)
	if segNum <= 0 {
		return nil, ParamIllegalError
	}
	if segSize < 1<<4 {
		return nil, SizeTooSmallError
	}
	segNum = mathutil.Min(getSegLen(segNum, log), 256)

	s := newStore(segSize, segNum, log)
	cache := &Cache{
		store: s,
		log:   log,
		k:     newKeyBuilder(),
	}
	log.Infof("cache init finished,total size = %d, segs = %d", size, segNum)
	return cache, nil
}

// Set write key, cacheValues, overdueTimestamp into cache.
// whether update or add cache, remove is first.
func (c *Cache) Set(key string, value Values, overdueTimestamp int64) error {
	var err error
	id := c.k.getKeyId(key, c.store.segNum)
	seg := c.store.segs[id]
	lock := c.store.locks[id]
	lock.Lock()
	defer lock.Unlock()
	// 1. remove old cache
	c.store.remove(id, key)
	// 2. set data into segment tmp
	err = seg.updatePair(key, value, time.Now().UnixNano()+overdueTimestamp)
	if err != nil {
		return err
	}
	// 3. add pair into segment
	err = c.store.write(id)
	//c.log.Infof("%v has add in cache", key)
	if err != nil {
		return err
	}
	return nil
}

// Remove remove cache
func (c *Cache) Remove(key string) (Values, error) {
	keyId := c.k.getKeyId(key, c.store.segNum)
	lock := c.store.locks[keyId]
	lock.Lock()
	defer lock.Unlock()
	remove, err := c.store.remove(keyId, key)
	if err != nil {
		return nil, err
	}
	return remove.value, nil
}

// Len return overdueTimestamp of key in cache
func (c *Cache) Len() int {
	cnt := 0
	for i, pair := range c.store.segs {
		c.store.locks[i].RLock()
		cnt += pair.length
		c.store.locks[i].RUnlock()
	}
	return cnt
}

// Get returns cache from key whether key is expired.
// nil will return if key dose not hit.
func (c *Cache) Get(key string) (Values, bool, error) {
	var (
		err       error
		keyId     int
		freshPair *pair
	)
	keyId = c.k.getKeyId(key, c.store.segNum)
	lock := c.store.locks[keyId]
	seg := c.store.segs[keyId]
	lock.Lock()
	defer lock.Unlock()
	if p, err := c.store.remove(keyId, key); err == nil {
		freshPair = p
	} else {
		return nil, false, nil
	}
	err = seg.updatePair(key, freshPair.value, freshPair.overdueTimestamp)
	if err != nil {
		return nil, false, err
	}
	err = c.store.write(keyId)
	if err != nil {
		return nil, false, err
	}
	if freshPair.overdueTimestamp < time.Now().UnixNano() {
		return freshPair.value, true, nil
	}
	return freshPair.value, false, nil
}

// IncreaseSize add specific size of max size
func (c *Cache) IncreaseSize(size int64) {
}

// DecrementSize reduce specific size of max size
func (c *Cache) DecrementSize(size int64) error {
	return nil
}

type keyBuilder struct {
	b   []byte
	mtx *sync.Mutex
}

func newKeyBuilder() *keyBuilder {
	return &keyBuilder{b: make([]byte, 1024), mtx: &sync.Mutex{}}
}

func (k *keyBuilder) getKeyId(str string, segNum int) int {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	for i := range str {
		k.b[i] = str[i]
	}
	return int(xxhash.Sum64(k.b[:len(str)]) & uint64(segNum-1))
}

func GenerateKey(keys []string) string {
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) > 0
	})
	return strings.Join(keys, "")
}

func MarshalValue(o interface{}) (Values, error) {
	d, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return Values{ByteSliceValue{
		value: d,
	}}, nil
}

func GetByteValue(d byte) (Values, error) {
	return Values{ByteValue{
		value: d,
	}}, nil
}

func GetIntValue(d int64) (Values, error) {
	return Values{IntValue{
		value: d,
	}}, nil
}

func GetUnsignedValue(d uint64) (Values, error) {
	return Values{UnsignedValue{
		value: d,
	}}, nil
}

func GetBoolValue(d bool) (Values, error) {
	return Values{BoolValue{
		value: d,
	}}, nil
}

func GetByteSliceValue(d []byte) (Values, error) {
	return Values{ByteSliceValue{
		value: d,
	}}, nil
}

func GetStringValue(d string) (Values, error) {
	return Values{StringValue{
		value: d,
	}}, nil
}
