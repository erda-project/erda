// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cache

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
)

var (
	EvictError              = errors.New("cache memory evicted")
	NilPtrError             = errors.New("nil ptr error")
	EntryTypeDifferentError = errors.New("entry and cacheValues type different")
	WrongEntryKeyError      = errors.New("entry key is wrong")
	InitEntryFailedError    = errors.New("entry initialized failed")
	ValueNotFoundError      = errors.New("cacheValues not found")
	KeyNotFoundError        = errors.New("key  not found")
	ValueTypeNotFoundError  = errors.New("cacheValues type not found")
	IllegalCacheSize        = errors.New("illegal cache size")
)

// Cache implement concurrent safe cache with LRU and ttl strategy.
type Cache struct {
	store *store
	log   *logrus.Logger
}

// store implement LRU strategy
type store struct {
	maxSize     int64
	used        int64
	cacheKeys   *cmpCacheKeys
	cacheValues *cmpCacheValues
	log         *logrus.Logger
	mtx         *sync.RWMutex
}
type keyPair struct {
	value  interface{}
	number int64
}

//func newKeyPair(key interface{}, number int64) *keyPair {
//
//	return &keyPair{
//		cacheValues:  key,
//		number: number,
//	}
//}
func newStore(maxSize int64, logger *logrus.Logger) *store {
	return &store{
		cacheKeys:   newCmpCacheKeys(),
		cacheValues: newCmpCacheValue(),
		log:         logger,
		maxSize:     maxSize,
		mtx:         &sync.RWMutex{},
	}
}
func (s *store) write(key string, freshEntry *entry) error {

	needSize, _ := freshEntry.getEntrySize()
	if s.maxSize < needSize {
		s.log.Errorf("evict cache size,try next")
		return EvictError
	}
	var (
		cacheEntry *entry
		tail       *list.Element
	)
	usage := s.used
	e, _ := s.cacheValues.value.Load(key)
	tail = s.cacheKeys.keys.Back()
	// 1. move key to tail. maybe new entry, handle error is unnecessary
	s.cacheKeys.moveToBack(key)

	if e != nil {
		cacheEntry = e.(*entry)
		cSize, _ := cacheEntry.getEntrySize()
		usage -= cSize
	}

	snapUsage := usage

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if tail != nil {
		// 2. sum memory could reduce when required by freshEntry not sufficient
		head := s.cacheKeys.keys.Front()
		for ; head != tail.Next(); head = head.Next() {
			if s.maxSize-usage < needSize {
				rEntry, _ := s.cacheValues.value.Load(head.Value.(*keyPair).value)
				rSize, _ := rEntry.(*entry).getEntrySize()
				usage -= rSize
			} else {
				break
			}
		}
		// 3. reduce memory from oldest cache
		if snapUsage != usage {
			for elem := s.cacheKeys.keys.Front(); elem != head; {
				cur := elem
				elem = elem.Next()
				rKey := cur.Value.(*keyPair).value
				s.remove(rKey, cur)
				s.log.Warnf("not enough memory,remove %v",rKey)
			}
		}
	}

	usage += needSize
	s.cacheValues.value.Store(key, freshEntry)
	s.used = usage
	s.cacheKeys.cnt++
	s.log.Infof("%v has add in cache", key)
	return nil
}

func (s *store) load(key string) (interface{}, error) {
	if v, ok := s.cacheValues.value.Load(key); !ok {
		return nil, ValueNotFoundError
	} else {
		s.mtx.Lock()
		defer s.mtx.Unlock()
		err := s.cacheKeys.moveToBack(key)
		if err != nil {
			return nil, err
		}
		return v, nil
	}
}

func (s *store) remove(key interface{}, rElem *list.Element) {
	delete(s.cacheKeys.mapping, key.(string))
	s.cacheKeys.keys.Remove(rElem)
	s.cacheKeys.pool.Put(rElem.Value)
	entry, _ := s.cacheValues.value.LoadAndDelete(key)
	s.cacheValues.pool.Put(entry)
}

// cmpCacheKeys store keys orderly. head elem is the smallest.
type cmpCacheKeys struct {
	keys    *list.List
	mapping map[string]*list.Element
	cnt     int64
	pool    *sync.Pool
}

type cmpCacheValues struct {
	value *sync.Map
	pool  *sync.Pool
}

func newCmpCacheKeys() *cmpCacheKeys {
	return &cmpCacheKeys{
		keys:    list.New(),
		cnt:     0,
		mapping: map[string]*list.Element{},
		pool: &sync.Pool{New: func() interface{} {
			return &keyPair{
				value:  nil,
				number: 0,
			}
		}},
	}
}

func newCmpCacheValue() *cmpCacheValues {
	return &cmpCacheValues{
		value: &sync.Map{},
		pool: &sync.Pool{New: func() interface{} {
			return &entry{
				serializable:   false,
				key:            "",
				startTimeStamp: 0,
				endTimeStamp:   0,
				value:          nil,
				entryType:      0,
			}
		}},
	}
}

//// SortKeys returns sorted cacheKeys by pair number
//func (c cmpCacheKeys) SortKeys() []*keyPair {
//	if c.keyPairs.Len() == 0 {
//		return nil
//	}
//	cacheKeys := make([]*keyPair, c.keyPairs.Len())
//	copy(cacheKeys, c.keyPairs.pairs)
//	sort.Slice(cacheKeys, c.keyPairs.Less)
//	return cacheKeys
//}

// moveToBack move latest keyPair to the end of KeyPairs
func (c cmpCacheKeys) moveToBack(key string) error {
	if elem, ok := c.mapping[key]; !ok {
		kp := c.pool.Get().(*keyPair)
		kp.number = c.cnt
		kp.value = key
		c.mapping[key] = c.keys.PushBack(kp)
		return WrongEntryKeyError
	} else {
		elem.Value.(*keyPair).number = c.cnt
		c.keys.MoveToBack(elem)
		return nil
	}
}

// Entry struct contains serializable data, always sorted by timestamp.
// Also can be used to store common data.
type entry struct {
	serializable   bool
	key            string
	startTimeStamp int64
	endTimeStamp   int64
	value          Values
	entryType      EntryType
	overdueTimestamp int64
}

//func newEntry() *entry {
//	return &entry{}
//}

type CmpCache interface {
	Remove(key string) error
	WriteMulti(serializable map[string]bool, pairs map[string]Values) error
	Write(serializable bool, key string, value Values) error
	IncreaseSize(size int64)
	DecrementSize(size int64) error
	Get(key string) (Values, error)
}

// newEntry return new entry.
func (s *store) updateEntry(serializable bool, key string, newValues Values, freshEntry *entry) error {
	length := len(newValues)
	if length == 0 {
		return InitEntryFailedError
	}
	entryType := returnValueType(newValues[0])
	if !freshEntry.serializable {
		entrySortByValue(newValues, entryType)
		deduplicateByValue(newValues)
	} else {
		entrySortByTimestamp(newValues)
		deduplicateByTimeStamp(newValues)
	}
	if freshEntry.key == "" && freshEntry.serializable && freshEntry.endTimeStamp >= newValues[length-1].TimeStamp() && freshEntry.startTimeStamp <= newValues[0].TimeStamp() {
		s.log.Infof("values is subset of exist entry,continue")
		return nil
	}
	freshEntry.serializable = serializable
	freshEntry.value = newValues
	freshEntry.key = key
	freshEntry.startTimeStamp = newValues[0].TimeStamp()
	freshEntry.endTimeStamp = freshEntry.value[len(freshEntry.value)-1].TimeStamp()
	return nil
}

// getEntrySize returns total size of values.
func (freshEntry *entry) getEntrySize() (int64, error) {
	if freshEntry == nil {
		return 0, NilPtrError
	}
	var usage = int64(0)
	for _, v := range freshEntry.value {
		usage += v.Size()
	}
	// 21 = sizeof int64 *2 + bool + int
	return usage , nil
}

func (freshEntry *entry) update(fresh *entry) {
	freshEntry.value = fresh.value
	freshEntry.serializable = fresh.serializable
	freshEntry.endTimeStamp = fresh.endTimeStamp
	freshEntry.startTimeStamp = fresh.startTimeStamp
	freshEntry.entryType = fresh.entryType
}

func entrySortByValue(value Values, entryType EntryType) {
	switch entryType {
	case IntType:
		sort.Slice(value, func(i, j int) bool {
			return value[i].(IntValue).value < value[j].(IntValue).value
		})
	case FloatType:
		sort.Slice(value, func(i, j int) bool {
			return value[i].(FloatValue).value < value[j].(FloatValue).value
		})
	case StringType:
		sort.Slice(value, func(i, j int) bool {
			return value[i].(StringValue).value < value[j].(StringValue).value
		})
	case UnsignedType:
		sort.Slice(value, func(i, j int) bool {
			return value[i].(UnsignedValue).value < value[j].(UnsignedValue).value
		})
	}
}

func entrySortByTimestamp(values Values) {
	sort.Slice(values, func(i, j int) bool {
		return values[i].TimeStamp() < values[j].TimeStamp()
	})
}

// deduplicate values must be sorted
func deduplicateByValue(value Values) {
	length := len(value)
	if length <= 1 {
		return
	}
	var val = value[length-1]
	for i := length - 2; i >= 0; i-- {
		if value[i].Value() == val.Value() {
			value = append(value[:i], value[i+1:]...)
		}
		val = value[i]
	}
}

// deduplicate values must be sorted
func deduplicateByTimeStamp(value Values) {
	length := len(value)
	if length <= 1 {
		return
	}
	var val = value[length-1]
	for i := length - 2; i >= 0; i-- {
		if value[i].TimeStamp() == val.TimeStamp() {
			value = append(value[:i], value[i+1:]...)
		}
		val = value[i]
	}
}

func returnValueType(v Value) EntryType {
	switch v.(type) {
	case FloatValue:
		return FloatType
	case IntValue:
		return IntType
	case BoolValue:
		return BoolType
	case UnsignedValue:
		return UnsignedType
	case StringValue:
		return StringType
	default:
		return UndefinedType
	}
}

type Value interface {
	// String returns string
	String() string
	// Type returns type of value
	Type() EntryType
	// TimeStamp returns value native timestamp or create timestamp
	TimeStamp() int64
	// Size returns size of value
	Size() int64
	// Value returns any type
	Value() interface{}
}

type FloatValue struct {
	unixnano int64
	value    float64
}

func (f FloatValue) Size() int64 {
	return 16
}

func (f FloatValue) TimeStamp() int64 {
	return f.unixnano
}

func (f FloatValue) String() string {
	return fmt.Sprintf("%f", f.value)
}

func (f FloatValue) Type() EntryType {
	return IntType
}

func (f FloatValue) Value() interface{} {
	return f.value
}

type IntValue struct {
	unixnano int64
	value    int64
}

func (i IntValue) Size() int64 {
	return 16
}

func (i IntValue) TimeStamp() int64 {
	return i.unixnano
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
	unixnano int64
	value    string
}

func (s StringValue) Size() int64 {
	return 8 + int64(len(s.value))
}

func (s StringValue) TimeStamp() int64 {
	return s.unixnano
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
	unixnano int64
	value    uint64
}

func (u UnsignedValue) Size() int64 {
	return 16
}

func (u UnsignedValue) TimeStamp() int64 {
	return u.unixnano
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
	unixnano int64
	value    bool
}

func (b BoolValue) Size() int64 {
	return 9
}

func (b BoolValue) TimeStamp() int64 {
	return b.unixnano
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
func New(size int64) *Cache {
	log := logrus.New()
	cache := &Cache{
		store: newStore(size, log),
		log:   log,
	}
	return cache
}

// WriteMulti write each key cacheValues pair into cache
func (c *Cache) WriteMulti(serializable map[string]bool, pairs map[string]Values) error {
	for k, v := range pairs {
		if serialize, ok := serializable[k]; !ok {
			return ValueTypeNotFoundError
		} else {
			err := c.Write(serialize, k, v)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Write write key cacheValues pair into cache
func (c *Cache) Write(serializable bool, key string, value Values) error {

	var (
		freshEntry *entry
		err        error
	)
	if v, err := c.store.load(key); err != nil {
		freshEntry = c.store.cacheValues.pool.Get().(*entry)
	} else {
		freshEntry = v.(*entry)
	}
	err = c.store.updateEntry(serializable, key, value, freshEntry)
	if err != nil {
		return err
	}
	err = c.store.write(key, freshEntry)
	if err != nil {
		return err
	}
	return nil
}

// Remove remove cache
func (c *Cache) Remove(key string) error {
	c.store.mtx.Lock()
	defer c.store.mtx.Unlock()
	c.store.remove(key, c.store.cacheKeys.mapping[key])
	return nil
}

// Len return number of key in cache
func (c *Cache) Len() int {
	c.store.mtx.RLock()
	defer c.store.mtx.RUnlock()
	return c.store.cacheKeys.keys.Len()
}

// Get returns cache from key
func (c *Cache) Get(key string) (Values, error) {
	v, err := c.store.load(key)
	if err != nil {
		return nil, err
	}
	return v.(*entry).value, nil
}

// IncreaseSize add specific size of max size
func (c *Cache) IncreaseSize(size int64) {
	atomic.AddInt64(&c.store.maxSize, size)
}

// DecrementSize reduce specific size of max size
func (c *Cache) DecrementSize(size int64) error {
	usage := c.store.used
	if usage < size {
		c.log.Errorf("cache size can not be decrease to %v", size)
		return IllegalCacheSize
	}
	atomic.AddInt64(&c.store.maxSize, -size)
	return nil
}

func (c *Cache) GenerateKey(keys []string) string {
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) > 0
	})
	return strings.Join(keys, "")
}
