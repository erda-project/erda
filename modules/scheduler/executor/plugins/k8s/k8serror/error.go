// Package k8serror encapsulates error of k8s object
package k8serror

import "errors"

var (
	// ErrNotFound indicates "not found" error
	ErrNotFound = errors.New("not found")
	// ErrInvalidParams indicates invalid param(s)
	ErrInvalidParams = errors.New("invalid params")
)

// NotFound return whether it is a "not found" error
func NotFound(err error) bool {
	return notFound(err)
}

func notFound(err error) bool {
	return err == ErrNotFound
}
