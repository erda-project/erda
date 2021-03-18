package nexus

type maven struct{}

func (m maven) GetFormat() RepositoryFormat { return RepositoryFormatMaven }

// hosted
type MavenHostedRepositoryCreateRequest struct {
	maven
	HostedRepositoryCreateRequest
	Maven RepositoryMavenConfig `json:"maven,omitempty"`
}
type MavenHostedRepositoryUpdateRequest MavenHostedRepositoryCreateRequest

// proxy
type MavenProxyRepositoryCreateRequest struct {
	maven
	ProxyRepositoryCreateRequest
	Maven RepositoryMavenConfig `json:"maven,omitempty"`
}
type MavenProxyRepositoryUpdateRequest MavenProxyRepositoryCreateRequest

// group
type MavenGroupRepositoryCreateRequest struct {
	maven
	GroupRepositoryCreateRequest
}
type MavenGroupRepositoryUpdateRequest MavenGroupRepositoryCreateRequest
