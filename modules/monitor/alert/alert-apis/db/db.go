package db

import (
	"github.com/jinzhu/gorm"
)

// DB .
type DB struct {
	*gorm.DB
	CustomizeAlert               CustomizeAlertDB
	CustomizeAlertRule           CustomizeAlertRuleDB
	CustomizeAlertNotifyTemplate CustomizeAlertNotifyTemplateDB
	Alert                        AlertDB
	AlertExpression              AlertExpressionDB
	AlertNotify                  AlertNotifyDB
	AlertNotifyTemplate          AlertNotifyTemplateDB
	AlertRule                    AlertRuleDB
	AlertRecord                  AlertRecordDB
}

// New .
func New(db *gorm.DB) *DB {
	return &DB{
		DB:                           db,
		CustomizeAlert:               CustomizeAlertDB{db},
		CustomizeAlertRule:           CustomizeAlertRuleDB{db},
		CustomizeAlertNotifyTemplate: CustomizeAlertNotifyTemplateDB{db},
		Alert:                        AlertDB{db},
		AlertExpression:              AlertExpressionDB{db},
		AlertNotify:                  AlertNotifyDB{db},
		AlertNotifyTemplate:          AlertNotifyTemplateDB{db},
		AlertRule:                    AlertRuleDB{db},
		AlertRecord:                  AlertRecordDB{db},
	}
}

// Begin .
func (db *DB) Begin() *DB {
	return New(db.DB.Begin())
}
