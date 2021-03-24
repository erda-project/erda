package enhancedqueue

func (eq *EnhancedQueue) inPending(key string) bool {
	return eq.pending.Get(key) != nil
}

func (eq *EnhancedQueue) inProcessing(key string) bool {
	return eq.processing.Get(key) != nil
}
