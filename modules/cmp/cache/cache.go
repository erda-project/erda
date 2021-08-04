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
	EntryTypeDifferentError = errors.New("entry and value type different")
	WrongEntryKeyError      = errors.New("entry key is wrong")
	InitEntryFailedError    = errors.New("entry initialized failed")
	ValueNotFoundError      = errors.New("value not found")
	ValueTypeNotFoundError  = errors.New("value type not found")
	IllegalCacheSize        = errors.New("illegal cache size")
)

// Cache implement concurrent safe cache with LRU strategy.
type Cache struct {
	store *store
	log   *logrus.Logger
}

// store implement LRU strategy
type store struct {
	maxSize int64
	used    int64
	keys    *cmpCacheHeap
	value   *sync.Map
	log     *logrus.Logger
	mtx     *sync.RWMutex
}
type keyPair struct {
	value  interface{}
	number int64
}

func newKeyPair(key interface{}, number int64) *keyPair {
	return &keyPair{
		value:  key,
		number: number,
	}
}
func newStore(maxSize int64, logger *logrus.Logger) *store {
	return &store{
		keys:    newCmpCacheHeap(),
		value:   &sync.Map{},
		log:     logger,
		maxSize: maxSize,
		mtx:     &sync.RWMutex{},
	}
}
func (s *store) write(key string, freshEntry *entry) error {
	needSize, _ := freshEntry.getEntrySize()
	if s.maxSize < needSize {
		s.log.Errorf("evict cache size,try another one")
		return EvictError
	}
	var (
		cacheEntry *entry
		isNew      = false
		i          = 0
	)
	usage := s.used
	e, _ := s.value.Load(key)

	if e != nil {
		cacheEntry = e.(*entry)
	} else {
		cacheEntry = newEntry()
		isNew = true
	}

	if !isNew && freshEntry.serializable && freshEntry.endTimeStamp <= cacheEntry.endTimeStamp && freshEntry.startTimeStamp >= cacheEntry.startTimeStamp {
		s.log.Infof("values is subset of exist entry,continue")
		return nil
	}

	if !isNew {
		cSize, _ := cacheEntry.getEntrySize()
		usage -= cSize
	}

	i = 0
	snapUsage := usage
	keys := s.keys.Keys()
	// 1. sum memory could reduce when required by freshEntry not sufficient
	for ; i < len(keys); i++ {
		if s.maxSize-usage < needSize && keys[i].value != key {
			rEntry, _ := s.value.Load(keys[i].value)
			rSize, _ := rEntry.(*entry).getEntrySize()
			usage -= rSize
		}
	}
	// 2. reduce memory from oldest cache
	if snapUsage != usage {
		for j := 0; j < i; j++ {
			s.remove(keys[j].value)
			s.log.Warnf("not enough memory,remove %v", keys[j].value)
		}
	}
	usage += needSize
	s.keys.keyPairs.pairs = append(s.keys.keyPairs.pairs, newKeyPair(key, s.keys.cnt))
	s.value.Store(key, freshEntry)
	s.used = usage
	s.keys.cnt++
	s.log.Infof("%v has add in cache", key)
	return nil
}

func (s *store) load(key interface{}) (interface{}, error) {
	if v, ok := s.value.Load(key); !ok {
		return nil, ValueNotFoundError
	} else {
		return v, nil
	}
}

func (s *store) remove(key interface{}) (interface{}, error) {
	for i, v := range s.keys.keyPairs.pairs {
		if v.value == key {
			s.keys.keyPairs.pairs = append(s.keys.keyPairs.pairs[:i], s.keys.keyPairs.pairs[i+1:]...)
			s.value.Delete(key)
			return key, nil
		}
	}
	return nil, ValueNotFoundError
}

type KeyPairs struct {
	pairs []*keyPair
}

func (k KeyPairs) Len() int {
	return len(k.pairs)
}

func (k KeyPairs) Less(i, j int) bool {
	return k.pairs[i].number > k.pairs[j].number
}

func (k KeyPairs) Swap(i, j int) {
	k.pairs[i], k.pairs[j] = k.pairs[j], k.pairs[i]
}

type cmpCacheHeap struct {
	keyPairs KeyPairs
	cnt      int64
}

func newCmpCacheHeap() *cmpCacheHeap {
	return &cmpCacheHeap{
		keyPairs: KeyPairs{pairs: make([]*keyPair, 0)},
		cnt:      0,
	}
}

// Keys returns sorted keys by pair number
func (c cmpCacheHeap) Keys() []*keyPair {
	if c.keyPairs.Len() == 0 {
		return nil
	}
	keys := make([]*keyPair, c.keyPairs.Len())
	copy(keys, c.keyPairs.pairs)
	sort.Slice(keys, c.keyPairs.Less)
	return keys
}

// Entry struct contains serializable data, always sorted by timestamp.
// Also can be used to store common data
type entry struct {
	serializable   bool
	key            string
	startTimeStamp int64
	endTimeStamp   int64
	value          Values
	entryType      EntryType
}

func newEntry() *entry {
	return &entry{}
}

type CmpCache interface {
	Remove(key string) error
	WriteMulti(serializable map[string]bool, pairs map[string]Values) error
	IncreaseSize(size int64)
	DecrementSize(size int64) error
	Get(key string) (Values, error)
}

// updateOrNew returns update values if entry exist or create new entry.
func (e *entry) updateOrNew(key string, serializable bool, newValues Values) (*entry, error) {
	freshEntry := newEntry()
	freshEntry.serializable = serializable
	freshEntry.value = newValues
	freshEntry.key = key
	freshEntry.entryType = returnValueType(newValues[0])
	e.entryType = newValues[0].Type()
	// e not exist in cache ,then return freshEntry
	if e == nil || e.key == "" {
		if len(newValues) == 0 {
			return nil, InitEntryFailedError
		}
		if !freshEntry.serializable {
			freshEntry.entrySort()
		} else {
			freshEntry.entrySortByTimestamp()
		}
		freshEntry.deduplicate()
		freshEntry.startTimeStamp = newValues[0].TimeStamp()
		freshEntry.endTimeStamp = freshEntry.value[len(freshEntry.value)-1].TimeStamp()
		return freshEntry, nil
	}
	if e.key != key {
		return nil, WrongEntryKeyError
	}
	// if new values are sub set of cache then nothing to update
	if e.serializable && freshEntry.serializable && e.startTimeStamp <= freshEntry.startTimeStamp && e.endTimeStamp >= freshEntry.endTimeStamp {
		return e, nil
	}
	e.update(freshEntry)
	return e, nil
}

// getEntrySize returns total size of values.
func (e *entry) getEntrySize() (int64, error) {
	if e == nil {
		return 0, NilPtrError
	}
	var usage = int64(0)
	for _, v := range e.value {
		usage += v.Size()
	}
	// 21 = sizeof int64 *2 + bool + int
	return usage + 21 + int64(len(e.key)), nil
}

func (e *entry) update(fresh *entry) {
	e.value = fresh.value
	e.serializable = fresh.serializable
	e.endTimeStamp = fresh.endTimeStamp
	e.startTimeStamp = fresh.startTimeStamp
	e.entryType = fresh.entryType
}

func (e *entry) entrySort() {
	switch e.entryType {
	case IntType:
		sort.Slice(e.value, func(i, j int) bool {
			return e.value[i].(IntValue).value < e.value[j].(IntValue).value
		})
	case FloatType:
		sort.Slice(e.value, func(i, j int) bool {
			return e.value[i].(FloatValue).value < e.value[j].(FloatValue).value
		})
	case StringType:
		sort.Slice(e.value, func(i, j int) bool {
			return e.value[i].(StringValue).value < e.value[j].(StringValue).value
		})
	case UnsignedType:
		sort.Slice(e.value, func(i, j int) bool {
			return e.value[i].(UnsignedValue).value < e.value[j].(UnsignedValue).value
		})
	}
}

func (e *entry) entrySortByTimestamp() {
	values := e.value
	sort.Slice(values, func(i, j int) bool {
		return values[i].TimeStamp() < values[j].TimeStamp()
	})
}

// deduplicate values must be sorted
func (e *entry) deduplicate() {
	length := len(e.value)
	if length <= 1 {
		return
	}
	var val = e.value[length-1]
	for i := length - 2; i >= 0; i-- {
		if e.value[i].TimeStamp() == val.TimeStamp() {
			e.value = append(e.value[:i], e.value[i+1:]...)
		}
		val = e.value[i]
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

// WriteMulti write each key value pair into cache
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

// Write write key value pair into cache
func (c *Cache) Write(serializable bool, key string, value Values) error {
	c.store.mtx.Lock()
	defer c.store.mtx.Unlock()
	var (
		cacheKey *entry
		err      error
	)
	if v, err := c.store.load(key); err != nil {
		cacheKey = newEntry()
	} else {
		cacheKey = v.(*entry)
	}
	freshEntry, err := cacheKey.updateOrNew(key, serializable, value)
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
	keys := c.store.keys.Keys()
	for _, k := range keys {
		if k.value == key {
			if _, err := c.store.remove(key); err != nil {
				return err
			}
			return nil
		}
	}
	return ValueNotFoundError
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
