package bundle

import (
	"testing"
)

func TestBundle_CreateOrUpdateIngress(t *testing.T) {
	// os.Setenv("HEPA_ADDR", "hepa.default.svc.cluster.local:8080")
	// defer func() {
	// 	os.Unsetenv("HEPA_ADDR")
	// }()
	// logrus.SetOutput(os.Stdout)
	// b := New(WithHepa())
	// path := "/123/$1"
	// err := b.CreateOrUpdateComponentIngress(apistructs.ComponentIngressUpdateRequest{
	// 	ComponentName: "hepa",
	// 	ComponentPort: 8080,
	// 	IngressName:   "hepa-test",
	// 	Routes: []apistructs.IngressRoute{
	// 		{
	// 			Domain: "hepatest.test.terminus.io",
	// 			Path:   "/(.*)",
	// 		},
	// 		{
	// 			Domain: "hepatest2.test.terminus.io",
	// 			Path:   "/(.*)",
	// 		},
	// 	},
	// 	RouteOptions: apistructs.RouteOptions{
	// 		RewritePath: &path,
	// 		UseRegex:    true,
	// 		Annotations: map[string]string{
	// 			"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	// 		},
	// 	},
	// })
	// require.NoError(t, err)
}

func TestBundle_CreateOrUpdateIngressNexusDocker(t *testing.T) {
	// os.Setenv("HEPA_ADDR", "hepa.default.svc.cluster.local:8080")
	// defer func() {
	// 	os.Unsetenv("HEPA_ADDR")
	// }()
	// logrus.SetOutput(os.Stdout)
	// b := New(WithHepa())
	// path := "/repository/docker-hosted-platform/$1"
	// err := b.CreateOrUpdateComponentIngress(apistructs.ComponentIngressUpdateRequest{
	// 	ComponentName: "addon-nexus",
	// 	ComponentPort: 8081,
	// 	IngressName:   "addon-nexus-docker-hosted-platform",
	// 	Routes: []apistructs.IngressRoute{
	// 		{
	// 			Domain: "docker-hosted-nexus-sys.dev.terminus.io",
	// 			Path:   "/(.*)",
	// 		},
	// 	},
	// 	RouteOptions: apistructs.RouteOptions{
	// 		RewritePath: &path,
	// 		UseRegex:    true,
	// 		Annotations: map[string]string{
	// 			"nginx.ingress.kubernetes.io/proxy-body-size": "0",
	// 		},
	// 	},
	// })
	// require.NoError(t, err)
}
