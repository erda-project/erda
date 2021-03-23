package sqllint_test

import (
	"testing"

	"github.com/erda-project/erda/pkg/sqllint"
)

const necessaryFieldLinterTest = `
CREATE TABLE IF NOT EXISTS dice_api_slas
(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key',
    created_at DATETIME COMMENT 'create time',
    updated_at DATETIME COMMENT 'update time',
    creator_id VARCHAR(191) COMMENT 'creator id',
    updater_id VARCHAR(191) COMMENT 'creator id',

    name       VARCHAR(191) COMMENT 'SLA name',
    approval   VARCHAR(16) COMMENT 'auto, manual',
    access_id  bigint COMMENT 'access id'
);

CREATE TABLE IF NOT EXISTS dice_api_slas
(
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT 'primary key',
    created_at DATETIME default CURRENT_TIMESTAMP COMMENT 'create time',
    updated_at DATETIME default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP COMMENT 'update time',
    creator_id VARCHAR(191) COMMENT 'creator id',
    updater_id VARCHAR(191) COMMENT 'creator id',

    name       VARCHAR(191) COMMENT 'SLA name',
    approval   VARCHAR(16) COMMENT 'auto, manual',
    access_id  bigint COMMENT 'access id'
);
`

func TestNewNecessaryFieldLinter(t *testing.T) {
	linter := sqllint.New(
		sqllint.NewIDExistsLinter, sqllint.NewIDTypeLinter, sqllint.NewIDIsPrimaryLinter,
		sqllint.NewCreatedAtExistsLinter, sqllint.NewCreatedAtTypeLinter, sqllint.NewCreatedAtDefaultValueLinter,
		sqllint.NewUpdatedAtExistsLinter, sqllint.NewUpdatedAtTypeLinter, sqllint.NewUpdatedAtDefaultValueLinter, sqllint.NewUpdatedAtOnUpdateLinter)
	if err := linter.Input([]byte(necessaryFieldLinterTest), "necessaryFieldLinterTest"); err != nil {
		t.Error(err)
	}
	errors := linter.Errors()
	t.Logf("errors: %v", errors)
}
