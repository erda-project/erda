package restclient

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
)

func GetDefaultConfig(apiPath string) *rest.Config {
	if apiPath == "" {
		apiPath = "/apis"
	}
	return &rest.Config{
		APIPath:     apiPath,
		QPS:         1000,
		Burst:       100,
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(1000, 100),
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		UserAgent: rest.DefaultKubernetesUserAgent(),
	}
}
