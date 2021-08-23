// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package org

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
)

// func TestShouldGetOrgByName(t *testing.T) {
// 	db, mock, err := sqlmock.New()
// 	require.NoError(t, err)
// 	connection, err := gorm.Open("mysql", db)
// 	require.NoError(t, err)
// 	client := &dao.DBClient{
// 		connection,
// 	}

// 	const sql = `SELECT * FROM "dice_org" WHERE (name = ?)`
// 	const sql1 = ` ORDER BY "dice_org"."id" ASC LIMIT 1`
// 	str := regexp.QuoteMeta(sql + sql1)
// 	mock.ExpectQuery(str).
// 		WithArgs("org1").
// 		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
// 			AddRow(1, "org1"))

// 	org, err := client.GetOrgByName("org1")
// 	require.NoError(t, err)
// 	assert.Equal(t, org.ID, 1)

// 	require.NoError(t, mock.ExpectationsWereMet())
// }

func TestGetOrgByDomainAndOrgName(t *testing.T) {
	o := &Org{}
	org := &model.Org{Name: "org0"}
	orgByDomain := monkey.PatchInstanceMethod(reflect.TypeOf(o), "GetOrgByDomain", func(_ *Org, domain string) (*model.Org, error) {
		if domain == "org0" {
			return org, nil
		} else {
			return nil, nil
		}
	})
	defer orgByDomain.Unpatch()
	db := &dao.DBClient{}
	orgByName := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetOrgByName", func(_ *dao.DBClient, orgName string) (*model.Org, error) {
		if orgName == "org0" {
			return org, nil
		} else {
			return nil, dao.ErrNotFoundOrg
		}
	})
	defer orgByName.Unpatch()

	res, err := o.GetOrgByDomainAndOrgName("org0", "")
	require.NoError(t, err)
	assert.Equal(t, org, res)
	res, err = o.GetOrgByDomainAndOrgName("org0", "org1")
	require.NoError(t, err)
	assert.Equal(t, org, res)
	res, err = o.GetOrgByDomainAndOrgName("org2", "org1")
	require.NoError(t, err)
	assert.Equal(t, (*model.Org)(nil), res)
}

func TestOrgNameRetriever(t *testing.T) {
	var domains = []string{"erda-org.erda.cloud", "buzz-org.app.terminus.io", "fuzz.com"}
	var domainRoots = []string{"erda.cloud", "app.terminus.io"}
	assert.Equal(t, "erda", orgNameRetriever(domains[0], domainRoots[0]))
	assert.Equal(t, "buzz", orgNameRetriever(domains[1], domainRoots[1]))
	assert.Equal(t, "", orgNameRetriever(domains[2], domainRoots[0]))
}
