package input

type Input interface {
	// block until stopped
	Start(handler Handler) error
	// block until stopped
	Stop() error
	Name() string
}
