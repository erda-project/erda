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

package algo

import "testing"

func TestSWRREqualWeights(t *testing.T) {
	rr := NewSmoothWeightedRR([]WeightedItem{{ID: "A", Weight: 1}, {ID: "B", Weight: 1}})
	seq := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		item, ok := rr.Next()
		if !ok {
			t.Fatalf("unexpected no item at %d", i)
		}
		seq = append(seq, item.ID)
	}
	// Expect near-alternating sequence
	for i := 0; i < len(seq); i++ {
		expect := "A"
		if i%2 == 1 {
			expect = "B"
		}
		if seq[i] != expect {
			t.Fatalf("unexpected sequence at %d: got %s expect %s (seq=%v)", i, seq[i], expect, seq)
		}
	}
}

func TestSWRRWeightedDistribution(t *testing.T) {
	rr := NewSmoothWeightedRR([]WeightedItem{{ID: "A", Weight: 1}, {ID: "B", Weight: 3}})
	countA, countB := 0, 0
	maxRun := 0
	last := ""
	run := 0
	total := 100
	for i := 0; i < total; i++ {
		item, ok := rr.Next()
		if !ok {
			t.Fatalf("unexpected no item at %d", i)
		}
		if item.ID == "A" {
			countA++
		} else if item.ID == "B" {
			countB++
		}
		if item.ID == last {
			run++
		} else {
			if run > maxRun {
				maxRun = run
			}
			run = 1
			last = item.ID
		}
	}
	if run > maxRun {
		maxRun = run
	}
	// Ratio should be close to 1:3
	if countA < 20 || countA > 30 {
		t.Fatalf("countA out of expected range: %d", countA)
	}
	if countB < 70 || countB > 80 {
		t.Fatalf("countB out of expected range: %d", countB)
	}
	// Smoothness: no very long runs (should not exceed 4 for 1:3 weights)
	if maxRun > 4 {
		t.Fatalf("sequence not smooth, maxRun=%d", maxRun)
	}
}

func TestSWRRUpdateItems(t *testing.T) {
	rr := NewSmoothWeightedRR([]WeightedItem{{ID: "A", Weight: 1}, {ID: "B", Weight: 1}})
	// Warm up
	for i := 0; i < 10; i++ {
		if _, ok := rr.Next(); !ok {
			t.Fatal("unexpected no item during warmup")
		}
	}

	// Update to 1:3
	rr.UpdateItems([]WeightedItem{{ID: "A", Weight: 1}, {ID: "B", Weight: 3}})
	countA, countB := 0, 0
	total := 100
	for i := 0; i < total; i++ {
		item, ok := rr.Next()
		if !ok {
			t.Fatalf("unexpected no item at %d after update", i)
		}
		if item.ID == "A" {
			countA++
		} else if item.ID == "B" {
			countB++
		}
	}

	if countA < 20 || countA > 30 {
		t.Fatalf("after update countA out of range: %d", countA)
	}
	if countB < 70 || countB > 80 {
		t.Fatalf("after update countB out of range: %d", countB)
	}
}
