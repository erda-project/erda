package storage

type Queue interface {
	Produce(partition, topic string, data []byte) error
	Consume(handler func(partition, topic string, data []byte)) error
}

type LogStorage interface {
	// TODO walkSavedLogs
}

type MetricStorage interface {
}
