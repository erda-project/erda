// Package asset API 资产
package assetsvc

type Service struct{}

type Option func(*Service)

func New(options ...Option) *Service {
	r := &Service{}
	for _, op := range options {
		op(r)
	}
	return r
}
