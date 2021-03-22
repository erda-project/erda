package pvolumes

import "path"

const (
	EnvKeyMesosFetcherURI = "MESOS_FETCHER_URI"
)

// MakeMesosFetcherURI4AliyunRegistrySecret 生成 DC/OS mesos 下的 fetcherURI，相当于 k8s secret for aliyun docker registry
func MakeMesosFetcherURI4AliyunRegistrySecret(mountPoint string) string {
	return "file://" + path.Clean(mountPoint+"/docker-registry-aliyun/password.tar.gz")
}
