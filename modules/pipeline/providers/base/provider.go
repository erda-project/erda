package base

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine"
	"github.com/erda-project/erda/modules/pipeline/providers/base/pipelinesvc"
	"github.com/erda-project/erda/modules/pipeline/services/actionagentsvc"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/services/permissionsvc"
	"github.com/erda-project/erda/modules/pipeline/services/pipelinecronsvc"
	"github.com/erda-project/erda/modules/pipeline/services/queuemanage"
	"github.com/erda-project/erda/modules/pkg/websocket"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type config struct {
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register     `autowired:"service-register"`
	MySQL    mysqlxorm.Interface    `autowired:"mysql-xorm"`
	Cms      cmspb.CmsServiceServer `autowired:"core.pipeline.cms.CmsServics"`

	// implements
	baseService *baseService
}

func (p *provider) Init(ctx servicehub.Context) error {
	baseService := &baseService{}
	baseService.p = p

	// TODO change to providers after depends-on services split into providers
	// bundle
	bdl := bundle.New(bundle.WithAllAvailableClients())
	// db
	legacyDbClient, err := dbclient.New()
	if err != nil {
		return err
	}
	// jsonstore
	js, err := jsonstore.New()
	if err != nil {
		return err
	}
	// etcd
	etcdctl, err := etcd.New()
	if err != nil {
		return err
	}
	// publisher
	publisher, err := websocket.NewPublisher()
	if err != nil {
		return err
	}

	baseService.svc = pipelinesvc.New(
		appsvc.New(bdl),
		crondsvc.New(legacyDbClient, bdl, js),
		actionagentsvc.New(legacyDbClient, bdl, js, etcdctl),
		extmarketsvc.New(bdl),
		pipelinecronsvc.New(legacyDbClient, crondsvc.New(legacyDbClient, bdl, js)),
		permissionsvc.New(bdl),
		queuemanage.New(queuemanage.WithDBClient(legacyDbClient)),
		p.MySQL,
		bdl,
		publisher,
		pipengine.New(legacyDbClient),
		js,
		etcdctl,
	)
	baseService.svc.WithCmsService(p.Cms)

	if p.Register != nil {
		pb.RegisterBaseServiceImp(p.Register, p.baseService)
	}

	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.pipeline.base.BaseService" || ctx.Type() == pb.BaseServiceServerType() || ctx.Type() == pb.BaseServiceHandlerType():
		return p.baseService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.pipeline.base", &servicehub.Spec{
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
