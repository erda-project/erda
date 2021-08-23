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
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"modernc.org/mathutil"
)

func TestCache_Init(t *testing.T) {
	_, err := New(1<<32, 1<<32+1)
	if err == nil {
		t.Fail()
	}
	_, err = New(1<<32, 1<<4-1)
	if err == nil {
		t.Fail()
	}
}

func TestCache_DecrementSize(t *testing.T) {
	cache, err := New(1<<32, 1<<22)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "DecrementTest",
			args: args{
				100,
			},
		},
		{
			name: "DecrementTest",

			args: args{
				100,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := cache.DecrementSize(tt.args.size)
			if err != nil {

			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	cache, err := New(256*1024, 24)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    Values
		wantErr bool
	}{
		{"Get_Test",

			args{"metrics1"},
			Values{IntValue{1}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set("metrics1", Values{IntValue{
				value: 1,
			}}, int64(1))

			got, _, err := cache.Get("metrics1")

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, IntValue{2})
			}

			got, _, err = cache.Get("metrics2")

			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, IntValue{2})
			}
		})
	}
}

func TestCache_IncreaseSize(t *testing.T) {
	cache, err := New(256*1024, 256)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
	}{
		{"IncreaseSize_Test",
			args{
				10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.IncreaseSize(tt.args.size)
		})
	}
}

func TestCache_Remove(t *testing.T) {
	cache, err := New(256*1024, 256)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "RemoveTest",
			args:    args{key: "metrics1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache.Set("metrics1", Values{IntValue{
				value: 0,
			}}, int64(1))
			if _, err := cache.Remove(tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_Write(t *testing.T) {
	cache, err := New(1024*256, 256)
	if err != nil {
		t.Fatal(err)
	}
	if cache == nil {
		t.Fail()
		return
	}
	type args struct {
		pairs map[string]Values
	}
	tests := []struct {
		name string

		args    args
		wantErr bool
	}{
		{
			name: "WriteTest",
			args: args{
				pairs: map[string]Values{
					"metricsInt": {
						IntValue{
							value: 0,
						},
						IntValue{
							value: 10,
						},
					},

					"metricsStr": {
						StringValue{
							value: "123123131",
						},
						StringValue{
							value: "3213123",
						},
						StringValue{
							value: "4121231",
						},
					},
					"metricsFloat": {
						FloatValue{
							value: 3.1415,
						},
						FloatValue{
							value: 3.32,
						},
					},
					"metricsUint": {
						UnsignedValue{
							value: ^uint64(0),
						},
						UnsignedValue{
							value: ^uint64(0) >> 1,
						},
					},
					"metricsBool": {
						BoolValue{
							value: true,
						},
						BoolValue{
							value: true,
						},
					},
					"metricsIntSeria": {
						IntValue{
							value: 0,
						}, IntValue{
							value: 2,
						},
					},
					"metricsStrSeria": {
						StringValue{
							value: "123123131",
						},
					},
					"metricsFloatSeria": {
						FloatValue{
							value: 3.1415,
						},
						FloatValue{
							value: 3.1414,
						},
					},
					"metricsUintSeria": {
						UnsignedValue{
							value: ^uint64(0),
						},
						UnsignedValue{
							value: 0,
						},
					},
					"metricsBoolSeria": {
						BoolValue{
							value: true,
						},
						BoolValue{
							value: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "UpdateTest",
			args: args{
				pairs: map[string]Values{
					"metricsInt": {IntValue{
						value: 1,
					}},
					"metricsStr": {
						StringValue{
							value: "31",
						},
					},
					"metricsFloat": {
						FloatValue{
							value: 3.52414124124,
						},
					},
					"metricsUint": {
						UnsignedValue{
							value: ^uint64(0) >> 1,
						},
					},
					"metricsIntSeria": {
						IntValue{
							value: 200,
						},
						IntValue{
							value: 200,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "WriteBigDataTest",
			args: args{

				pairs: map[string]Values{

					"metricsStr": {
						StringValue{
							value: string(make([]byte, 1024*1024)),
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.args.pairs {
				if err := cache.Set(k, v, time.Now().UnixNano()); (err != nil) != tt.wantErr {
					t.Errorf("WriteMulti() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

		})

	}
}

func BenchmarkLRU_Rand(b *testing.B) {
	l, err := New(256*1024, 256)
	if err != nil {
		b.Fatal(err)
	}
	l.log.SetLevel(logrus.ErrorLevel)
	trace := make([]string, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = fmt.Sprintf("%d", rand.Int63()%32768)
	}

	b.ResetTimer()
	v := Values{IntValue{1}}
	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Set(trace[i], v, mathutil.MaxInt)
		} else {
			_, _, err := l.Get(trace[i])
			if err != nil {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func BenchmarkLRU_Freq1(b *testing.B) {
	c, err := New(1024*1024, 256)
	if err != nil {
		b.Fail()
		return
	}
	c.log.SetLevel(logrus.ErrorLevel)
	trace := make([]string, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = fmt.Sprintf("%d", rand.Int63()%16384)
		} else {
			trace[i] = fmt.Sprintf("%d", rand.Int63()%32768)
		}
	}

	v := Values{IntValue{1}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Set(trace[i], v, mathutil.MaxInt)
	}
	var hit, miss int

	for i := 0; i < b.N; i++ {
		_, _, err := c.Get(trace[i])
		if err != nil {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(miss)/float64(hit))
}

func BenchmarkLRU_FreqParallel(b *testing.B) {
	c, err := New(256*1024, 256)
	if err != nil {
		b.Fatal(err)
	}
	c.log.SetLevel(logrus.ErrorLevel)
	trace := make([]string, b.N*2)
	v := Values{IntValue{1}}
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = fmt.Sprintf("%d", rand.Int63()%16384)
		} else {
			trace[i] = fmt.Sprintf("%d", rand.Int63()%32768)
		}
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		counter := 0
		b.ReportAllocs()
		for pb.Next() {
			c.Set(trace[counter], v, int64(counter))
			counter = counter + 1
			if counter > b.N {
				counter = 0
			}
		}
	})
}

func TestLRU(t *testing.T) {
	c, err := New(256*1024, 255)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1024; i++ {
		err := c.Set(fmt.Sprintf("%d", i), Values{IntValue{
			value: int64(i),
		}}, int64(i))
		if err != nil {
			t.Fatalf("cache insert error %v", err)
		}
	}
	if c.Len() != 1024 {
		t.Fatalf("bad len: %v", c.Len())
	}

	for i := 0; i < 128; i++ {
		if v, _, ok := c.Get(fmt.Sprintf("%d", i+128)); ok != nil || int(v[0].(IntValue).value) != i+128 {
			t.Fatalf("bad key: %v", i+128)
		}
	}
	for i := 128; i < 256; i++ {
		_, _, err := c.Get(fmt.Sprintf("%d", i))
		if err != nil {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		_, err := c.Remove(fmt.Sprintf("%d", i))
		if err != nil {
			t.Fatalf("should be deleted")
		}
		v, _, err := c.Get(fmt.Sprintf("%d", i))
		if v != nil {
			t.Fatalf("should be deleted")
		}
	}

	_, _, err = c.Get(fmt.Sprintf("%d", 192))
	if err != nil {
		return
	} // expect 192 to be last key in c.Keys()

	if c.Len() != 960 {
		t.Fatalf("bad len: %v", c.Len())
	}
	if _, _, ok := c.Get(fmt.Sprintf("%d", 960)); ok != nil {
		t.Fatalf("should contain nothing")
	}

	//ma := make(map[int][]string)
	for i := 0; i < 10240; i++ {
		k := fmt.Sprintf("%d", i)
		err := c.Set(k, Values{IntValue{
			value: int64(i),
		}}, int64(i))
		//slice := ma[int(xxhash.Sum64([]byte(k))&uint64(c.store.segNum-1))]
		//slice = append(slice, k)
		if err != nil {
			t.Fatalf("cache insert error %v", err)
		}
	}
	for i := 0; i < 10240; i++ {
		v, _, _ := c.Get(fmt.Sprintf("%d", i))
		if v != nil && int(v[0].(IntValue).value) != i {
			t.Fatalf("bad key: %v", i)
		}
	}
}
