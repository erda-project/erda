package db

import "github.com/jinzhu/gorm"

// AlertNotifyTemplateDB .
type AlertNotifyTemplateDB struct {
	*gorm.DB
}

// QueryEnabledByTypesAndIndexes .
func (db *AlertNotifyTemplateDB) QueryEnabledByTypesAndIndexes(
	types, indexes []string) ([]*AlertNotifyTemplate, error) {
	var templates []*AlertNotifyTemplate
	if err := db.
		Where("alert_type IN (?)", types).
		Where("alert_index IN (?)", indexes).
		Where("enable=?", true).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}
