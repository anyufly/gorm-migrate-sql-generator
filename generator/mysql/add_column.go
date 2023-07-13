package mysql

import (
	"fmt"
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (sqlGenerator *migrateSQLGenerator) AddColumn(execTx *gorm.DB, value interface{}, name string) (*result.SQLForTable, error) {
	var sql *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// avoid using the same name field
		f := stmt.Schema.LookUpField(name)
		if f == nil {
			return fmt.Errorf("failed to look up field with name: %s", name)
		}

		if !f.IgnoreMigration {
			executedTx := m.DB.Exec(
				"ALTER TABLE ? ADD ? ?",
				m.CurrentTable(stmt), clause.Column{Name: f.DBName}, m.DB.Migrator().FullDataTypeOf(f),
			)

			if executedTx.Error != nil {
				return executedTx.Error
			}

			sql = loadStmtSQL(executedTx, stmt.Table)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return sql, nil
}
