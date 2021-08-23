package accesskey

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/services/accesskey/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
	"github.com/jinzhu/gorm"
)

type config struct {
}

// +provider
type provider struct {
	Cfg              *config
	Log              logs.Logger
	Register         transport.Register
	accessKeyService *accessKeyService
	dao              *dao
	Perm             perm.Interface `autowired:"permission"`
	DB               *gorm.DB       `autowired:"mysql-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.dao = &dao{db: p.DB}
	p.accessKeyService = &accessKeyService{p}
	if p.Register != nil {
		pb.RegisterAccessKeyServiceImp(p.Register, p.accessKeyService, apis.Options(), p.Perm.Check(
			perm.NoPermMethod(pb.AccessKeyServiceServer.QueryAccessKeys),
			perm.NoPermMethod(pb.AccessKeyServiceClient.GetAccessKey),
			perm.NoPermMethod(pb.AccessKeyServiceClient.CreateAccessKeys),
			perm.NoPermMethod(pb.AccessKeyServiceClient.UpdateAccessKeys),
			perm.NoPermMethod(pb.AccessKeyServiceClient.DeleteAccessKeys),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.services.accesskey.AccessKeyService" || ctx.Type() == pb.AccessKeyServiceServerType() || ctx.Type() == pb.AccessKeyServiceHandlerType():
		return p.accessKeyService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.services.accesskey", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
