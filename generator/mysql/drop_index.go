package mysql

import (
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (sqlGenerator *migrateSQLGenerator) DropIndex(execTx *gorm.DB, value interface{}, name string) (*generator.SQLForTable, error) {
	var sql *generator.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		executedTx := m.DB.Exec("DROP INDEX ? ON ?", clause.Column{Name: name}, m.CurrentTable(stmt))

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
