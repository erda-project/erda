package exporter

// Output .
type Output interface {
	Write(key string, data []byte) error
}

// OutputFactory .
type OutputFactory func(i int) (Output, error)

// Interface .
type Interface interface {
	NewConsumer(OutputFactory) error
}
