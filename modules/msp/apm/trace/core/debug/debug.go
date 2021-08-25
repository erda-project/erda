package debug

type Status int32

const (
	Init    Status = 0
	Success Status = 1
	Fail    Status = 2
	Stop    Status = 3
)
