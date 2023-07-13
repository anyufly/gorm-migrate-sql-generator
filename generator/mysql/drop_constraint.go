package mysql

import (
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (sqlGenerator *migrateSQLGenerator) DropConstraint(execTx *gorm.DB, value interface{}, name string) (*generator.SQLForTable, error) {
	var sqlForTable *generator.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {

		constraint, chk, table := m.GuessConstraintAndTable(stmt, name)
		if chk != nil {
			executedTx := m.DB.Exec("ALTER TABLE ? DROP CHECK ?", clause.Table{Name: stmt.Table}, clause.Column{Name: chk.Name})
			if executedTx.Error != nil {
				return executedTx.Error
			}

			sqlForTable = loadStmtSQL(executedTx, stmt.Table)
			return nil
		}

		if constraint != nil {
			name = constraint.Name
		}

		executedTx := m.DB.Exec(
			"ALTER TABLE ? DROP FOREIGN KEY ?", clause.Table{Name: table}, clause.Column{Name: name},
		)
		if executedTx.Error != nil {
			return executedTx.Error
		}
		sqlForTable = loadStmtSQL(executedTx, stmt.Table)

		return nil
	}); err != nil {
		return nil, err
	}

	return sqlForTable, nil
}
