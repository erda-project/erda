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

package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

// AlertExpressionDB .
type AlertExpressionDB struct {
	*gorm.DB
}

// QueryByIDs .
func (db *AlertExpressionDB) QueryByIDs(ids []uint64) ([]*AlertExpression, error) {
	var expressions []*AlertExpression
	if err := db.
		Where("id IN (?)", ids).
		Find(&expressions).Error; err != nil {
		return nil, err
	}

	return expressions, nil
}

// QueryByAlertIDs .
func (db *AlertExpressionDB) QueryByAlertIDs(alertIDs []uint64) ([]*AlertExpression, error) {
	var expressions []*AlertExpression
	if err := db.Where("alert_id IN (?)", alertIDs).Find(&expressions).Error; err != nil {
		return nil, err
	}
	return expressions, nil
}

// Insert .
func (db *AlertExpressionDB) Insert(expression *AlertExpression) error {
	expression.Created = time.Now()
	expression.Updated = time.Now()
	return db.Create(&expression).Error
}

// Update .
func (db *AlertExpressionDB) Update(expression *AlertExpression) error {
	expression.Updated = time.Now()
	return db.Model(&expression).Updates(&expression).Error
}

// UpdateEnableByAlertID .
func (db *AlertExpressionDB) UpdateEnableByAlertID(alertID uint64, enable bool) error {
	return db.
		Model(&AlertExpression{}).
		Where("alert_id=?", alertID).
		Update("updated", time.Now()).
		Update("enable", enable).Error
}

// DeleteByAlertID .
func (db *AlertExpressionDB) DeleteByAlertID(alertID uint64) error {
	return db.Where("alert_id=?", alertID).Delete(AlertExpression{}).Error
}

// DeleteByIDs .
func (db *AlertExpressionDB) DeleteByIDs(ids []uint64) error {
	return db.Where("id IN (?)", ids).Delete(AlertExpression{}).Error
}
