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

package main

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	TableAlertExpression = "sp_alert_expression"
	TableAlertRules      = "sp_alert_rules"
)

// AlertExpression .
type AlertExpression struct {
	ID         uint64        `gorm:"column:id"`
	AlertID    uint64        `gorm:"column:alert_id"`
	Attributes utils.JSONMap `gorm:"column:attributes"`
	Expression utils.JSONMap `gorm:"column:expression"`
	Version    string        `gorm:"column:version"`
	Enable     bool          `gorm:"column:enable"`
	Created    time.Time     `gorm:"column:created"`
	Updated    time.Time     `gorm:"column:updated"`
}

// TableName 。
func (AlertExpression) TableName() string { return TableAlertExpression }

// AlertRule .
type AlertRule struct {
	ID         uint64        `gorm:"column:id"`
	Name       string        `gorm:"column:name"`
	AlertScope string        `gorm:"column:alert_scope"`
	AlertType  string        `gorm:"column:alert_type"`
	AlertIndex string        `gorm:"column:alert_index"`
	Template   utils.JSONMap `gorm:"column:template"`
	Attributes utils.JSONMap `gorm:"column:attributes"`
	Version    string        `gorm:"column:version"`
	Enable     bool          `gorm:"column:enable"`
	CreateTime time.Time     `gorm:"column:create_time"`
	UpdateTime time.Time     `gorm:"column:update_time"`
}

// TableName .
func (AlertRule) TableName() string { return TableAlertRules }

func NewDBConn(c *config) *gorm.DB {
	db, err := gorm.Open("mysql", url(c))
	if err != nil {
		panic(fmt.Sprintf("fail to connect mysql: %s", err))
	}
	return db
}

type config struct {
	MySQLURL      string
	MySQLHost     string
	MySQLPort     string
	MySQLUsername string
	MySQLPassword string
	MySQLDatabase string
}

func url(c *config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		c.MySQLUsername, c.MySQLPassword, c.MySQLHost, c.MySQLPort, c.MySQLDatabase)
}

func migrateAlertExpression(db *gorm.DB) error {
	selectMap, err := createIndexRuleMap(db)
	if err != nil {
		return errors.Wrap(err, "createIndexRuleMap failed")
	}

	newExps, err := updatedSelectField(selectMap, db)
	if err != nil {
		return errors.Wrap(err, "updatedSelectField failed")
	}

	err = save2db(newExps, db)
	if err != nil {
		return errors.Wrap(err, "save2db failed")
	}
	return nil
}

func save2db(exps []*AlertExpression, db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var cnt int
		for _, item := range exps {
			if err := tx.Save(item).Error; err != nil {
				return err
			}
			cnt++
		}
		fmt.Printf("%d rows affected\n", cnt)
		return nil
	})
}

func updatedSelectField(selectMap map[string]*AlertRule, db *gorm.DB) ([]*AlertExpression, error) {
	var exps []*AlertExpression
	if err := db.Where(&AlertExpression{Enable: true}).Find(&exps).Error; err != nil {
		return nil, err
	}

	res := make([]*AlertExpression, 0, len(exps))
	for _, item := range exps {
		index, ok := item.Attributes["alert_index"]
		if !ok {
			continue
		}
		rule, ok := selectMap[index.(string)]
		if !ok {
			continue
		}
		dst, ok := rule.Template["select"]
		if !ok {
			continue
		}
		item.Expression["select"] = dst
		res = append(res, item)
	}
	return res, nil
}

func createIndexRuleMap(db *gorm.DB) (map[string]*AlertRule, error) {
	var rules []*AlertRule
	if err := db.Where(&AlertRule{Enable: true}).Find(&rules).Error; err != nil {
		return nil, err
	}
	res := make(map[string]*AlertRule)
	for _, item := range rules {
		res[item.AlertIndex] = item
	}
	return res, nil
}
