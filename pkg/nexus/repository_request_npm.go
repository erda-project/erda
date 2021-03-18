package nexus

type npm struct{}

func (n npm) GetFormat() RepositoryFormat { return RepositoryFormatNpm }

// hosted
type NpmHostedRepositoryCreateRequest struct {
	npm
	HostedRepositoryCreateRequest
}
type NpmHostedRepositoryUpdateRequest NpmHostedRepositoryCreateRequest

// proxy
type NpmProxyRepositoryCreateRequest struct {
	npm
	ProxyRepositoryCreateRequest
}
type NpmProxyRepositoryUpdateRequest NpmProxyRepositoryCreateRequest

// group
type NpmGroupRepositoryCreateRequest struct {
	npm
	GroupRepositoryCreateRequest
}
type NpmGroupRepositoryUpdateRequest NpmGroupRepositoryCreateRequest
