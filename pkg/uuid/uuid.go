package uuid

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Generate 不要再调用这个函数，太丑了，找时间废除.
func Generate() string {
	u := uuid.NewV4()
	return fmt.Sprintf("%x%x%x%x%x", u[:4], u[4:6], u[6:8], u[8:10], u[10:])
}

// UUID 返回 uuid.
func UUID() string {
	return Generate()
}
