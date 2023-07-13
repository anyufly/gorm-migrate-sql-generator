package mysql

import (
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (sqlGenerator *migrateSQLGenerator) CreateConstraint(execTx *gorm.DB, value interface{}, name string) (*result.SQLForTable, error) {
	var sqlForTable *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, chk, table := m.GuessConstraintAndTable(stmt, name)
		if chk != nil {
			executedTx := m.DB.Exec(
				"ALTER TABLE ? ADD CONSTRAINT ? CHECK (?)",
				m.CurrentTable(stmt), clause.Column{Name: chk.Name}, clause.Expr{SQL: chk.Constraint},
			)

			if executedTx.Error != nil {
				return executedTx.Error
			}

			sqlForTable = loadStmtSQL(executedTx, stmt.Table)
		}

		if constraint != nil {
			vars := []interface{}{clause.Table{Name: table}}
			if stmt.TableExpr != nil {
				vars[0] = stmt.TableExpr
			}
			sql, values := buildConstraint(constraint)
			executedTx := m.DB.Exec("ALTER TABLE ? ADD "+sql, append(vars, values...)...)

			if executedTx.Error != nil {
				return executedTx.Error
			}

			sqlForTable = loadStmtSQL(executedTx, stmt.Table)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return sqlForTable, nil
}
