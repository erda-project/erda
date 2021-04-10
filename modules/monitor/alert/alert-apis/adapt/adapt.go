package adapt

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	//"terminus.io/dice/monitor/modules/domain/metrics/metricq"
	//block "terminus.io/dice/monitor/modules/business/dashboard/chart-block"
)

const (
	orgScope = "org"

	orgCustomizeType = "org_customize"
)

var (
	// ErrorAlreadyExists .
	ErrorAlreadyExists = fmt.Errorf("alert already exists")
)

type invalidParameterError struct {
	text string
}

// Error
func (e *invalidParameterError) Error() string { return e.text }

func invalidParameter(format string, args ...interface{}) error {
	return &invalidParameterError{
		text: fmt.Sprintf(format, args...),
	}
}

// IsInvalidParameterError .
func IsInvalidParameterError(err error) bool {
	_, ok := err.(*invalidParameterError)
	return ok
}

// IsAlreadyExistsError .
func IsAlreadyExistsError(err error) bool {
	return err == ErrorAlreadyExists
}

// Adapt .
type Adapt struct {
	l                      logs.Logger
	metricq                metricq.Queryer
	t                      i18n.Translator
	db                     *db.DB
	cql                    *cql.Cql
	bdl                    *bundle.Bundle
	cmdb                   *cmdb.Cmdb
	dashboardAPI           block.DashboardAPI
	orgFilterTags          map[string]bool
	microServiceFilterTags map[string]bool
	silencePolicies        map[string]bool
}

// New .
func New(
	l logs.Logger,
	metricq metricq.Queryer,
	t i18n.Translator,
	db *db.DB,
	cql *cql.Cql,
	bdl *bundle.Bundle,
	cmdb *cmdb.Cmdb,
	dashapi block.DashboardAPI,
	orgFilterTags map[string]bool,
	microServiceFilterTags map[string]bool,
	silencePolicies map[string]bool,
) *Adapt {
	return &Adapt{
		l:                      l,
		metricq:                metricq,
		t:                      t,
		db:                     db,
		bdl:                    bdl,
		cmdb:                   cmdb,
		cql:                    cql,
		orgFilterTags:          orgFilterTags,
		microServiceFilterTags: microServiceFilterTags,
		silencePolicies:        silencePolicies,
		dashboardAPI:           dashapi,
	}
}
