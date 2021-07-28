// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package adapt

import (
	"fmt"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/core/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	block "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
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
	l                           logs.Logger
	metricq                     metricq.Queryer
	t                           i18n.Translator
	db                          *db.DB
	cql                         *cql.Cql
	bdl                         *bundle.Bundle
	cmdb                        *cmdb.Cmdb
	dashboardAPI                block.DashboardAPI
	orgFilterTags               map[string]bool
	microServiceFilterTags      map[string]bool
	microServiceOtherFilterTags map[string]bool
	silencePolicies             map[string]bool
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
	microServiceOtherFilterTags map[string]bool,
	silencePolicies map[string]bool,
) *Adapt {
	return &Adapt{
		l:                           l,
		metricq:                     metricq,
		t:                           t,
		db:                          db,
		bdl:                         bdl,
		cmdb:                        cmdb,
		cql:                         cql,
		orgFilterTags:               orgFilterTags,
		microServiceFilterTags:      microServiceFilterTags,
		microServiceOtherFilterTags: microServiceOtherFilterTags,
		silencePolicies:             silencePolicies,
		dashboardAPI:                dashapi,
	}
}
