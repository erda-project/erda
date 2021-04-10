package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

// AlertNotifyDB .
type AlertNotifyDB struct {
	*gorm.DB
}

// QueryByAlertIDs .
func (db *AlertNotifyDB) QueryByAlertIDs(alertIDs []uint64) ([]*AlertNotify, error) {
	var notifies []*AlertNotify
	if err := db.Where("alert_id IN (?)", alertIDs).Find(&notifies).Error; err != nil {
		return nil, err
	}
	return notifies, nil
}

// Insert .
func (db *AlertNotifyDB) Insert(notify *AlertNotify) error {
	notify.Created = time.Now()
	notify.Updated = time.Now()
	return db.Create(&notify).Error
}

// Update .
func (db *AlertNotifyDB) Update(notify *AlertNotify) error {
	notify.Updated = time.Now()
	return db.Model(&notify).Updates(&notify).Error
}

// UpdateEnableByAlertID .
func (db *AlertNotifyDB) UpdateEnableByAlertID(alertID uint64, enable bool) error {
	return db.
		Model(&AlertNotify{}).
		Where("alert_id=?", alertID).
		Update("updated", time.Now()).
		Update("enable", enable).Error
}

// DeleteByAlertID .
func (db *AlertNotifyDB) DeleteByAlertID(alertID uint64) error {
	return db.Where("alert_id=?", alertID).Delete(AlertNotify{}).Error
}

// DeleteByIDs .
func (db *AlertNotifyDB) DeleteByIDs(ids []uint64) error {
	return db.Where("id IN (?)", ids).Delete(AlertNotify{}).Error
}
