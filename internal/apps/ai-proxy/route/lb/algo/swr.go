package algo

import "sync"

// WeightedItem represents an item with a relative weight.
type WeightedItem struct {
	ID     string
	Weight int
}

// SmoothWeightedRR implements Smooth Weighted Round Robin (SWRR).
// It is thread-safe.
type SmoothWeightedRR struct {
	mu            sync.Mutex
	items         []WeightedItem
	currentWeight []int
	totalWeight   int
}

func NewSmoothWeightedRR(items []WeightedItem) *SmoothWeightedRR {
	r := &SmoothWeightedRR{}
	r.UpdateItems(items)
	return r
}

// UpdateItems replaces the item list and resets internal state.
func (s *SmoothWeightedRR) UpdateItems(items []WeightedItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make([]WeightedItem, 0, len(items))
	s.currentWeight = make([]int, len(items))
	s.totalWeight = 0
	for _, it := range items {
		w := it.Weight
		if w <= 0 {
			w = 1
		}
		s.items = append(s.items, WeightedItem{ID: it.ID, Weight: w})
		s.totalWeight += w
	}
}

// Items returns a copy of current items for comparison/debugging.
func (s *SmoothWeightedRR) Items() []WeightedItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]WeightedItem, len(s.items))
	copy(cp, s.items)
	return cp
}

// Next returns next item according to SWRR. ok=false when no items.
func (s *SmoothWeightedRR) Next() (WeightedItem, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.items) == 0 || s.totalWeight == 0 {
		return WeightedItem{}, false
	}

	// Increase current weights.
	for i := range s.items {
		s.currentWeight[i] += s.items[i].Weight
	}

	// Find max current weight.
	maxIdx := 0
	maxWeight := s.currentWeight[0]
	for i := 1; i < len(s.currentWeight); i++ {
		if s.currentWeight[i] > maxWeight {
			maxWeight = s.currentWeight[i]
			maxIdx = i
		}
	}

	// Adjust selected weight.
	s.currentWeight[maxIdx] -= s.totalWeight

	return s.items[maxIdx], true
}
