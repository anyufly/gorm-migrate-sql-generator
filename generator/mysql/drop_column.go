package mysql

import (
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DropColumn drop value's `name` column
func (sqlGenerator *migrateSQLGenerator) DropColumn(execTx *gorm.DB, value interface{}, name string) (*result.SQLForTable, error) {
	var sql *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(name); field != nil {
			name = field.DBName
		}

		executedTx := m.DB.Exec(
			"ALTER TABLE ? DROP COLUMN ?", m.CurrentTable(stmt), clause.Column{Name: name},
		)

		if executedTx.Error != nil {
			return executedTx.Error
		}

		sql = loadStmtSQL(executedTx, stmt.Table)
		return nil
	}); err != nil {
		return nil, err
	}

	return sql, nil
}
