package mysql

import (
	"fmt"
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
)

func (sqlGenerator *migrateSQLGenerator) CreateIndex(execTx *gorm.DB, value interface{}, name string) (*result.SQLForTable, error) {
	var sql *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			opts := m.DB.Migrator().(migrator.BuildIndexOptionsInterface).BuildIndexOptions(idx.Fields, stmt)
			values := []interface{}{clause.Column{Name: idx.Name}, m.CurrentTable(stmt), opts}

			createIndexSQL := "CREATE "
			if idx.Class != "" {
				createIndexSQL += idx.Class + " "
			}
			createIndexSQL += "INDEX ? ON ??"

			if idx.Type != "" {
				createIndexSQL += " USING " + idx.Type
			}

			if idx.Comment != "" {
				createIndexSQL += fmt.Sprintf(" COMMENT '%s'", idx.Comment)
			}

			if idx.Option != "" {
				createIndexSQL += " " + idx.Option
			}

			executedTx := m.DB.Exec(createIndexSQL, values...)
			if executedTx.Error != nil {
				return executedTx.Error
			}

			sql = loadStmtSQL(executedTx, stmt.Table)
			return nil

		}

		return fmt.Errorf("failed to create index with name %s", name)
	}); err != nil {
		return nil, err
	}
	return sql, nil
}
