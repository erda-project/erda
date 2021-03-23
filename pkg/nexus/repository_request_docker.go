package nexus

type docker struct{}

func (d docker) GetFormat() RepositoryFormat { return RepositoryFormatDocker }

// hosted
type DockerHostedRepositoryCreateRequest struct {
	docker
	HostedRepositoryCreateRequest
	Docker RepositoryDockerConfig `json:"docker,omitempty"`
}
type DockerHostedRepositoryUpdateRequest DockerHostedRepositoryCreateRequest

// proxy
type DockerProxyRepositoryCreateRequest struct {
	docker
	ProxyRepositoryCreateRequest
	Docker      RepositoryDockerConfig      `json:"docker,omitempty"`
	DockerProxy RepositoryDockerProxyConfig `json:"dockerProxy,omitempty"`
}
type DockerProxyRepositoryUpdateRequest DockerProxyRepositoryCreateRequest
