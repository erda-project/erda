package report

import "fmt"

type reportWrite struct {
	capacity int
}

func NewReportWrite(capacity int) *reportWrite {
	return &reportWrite{
		capacity: capacity,
	}
}

func (w *reportWrite) Write(data interface{}) error {
	return nil
}
func (w *reportWrite) WriteN(data ...interface{}) (int, error) {
	if w.capacity <= 0 {
		err := fmt.Errorf("buffer max capacity")
		return 0, err
	}
	w.capacity -= len(data)
	return len(data), nil
}
func (w reportWrite) Close() error {
	return nil
}
