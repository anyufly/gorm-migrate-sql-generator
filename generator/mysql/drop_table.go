package mysql

import (
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/gorm"
)

func (sqlGenerator *migrateSQLGenerator) DropTable(execTx *gorm.DB, values ...interface{}) ([]*generator.SQLForTable, error) {
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	sqlList := make([]*generator.SQLForTable, 0, 10)

	values = m.ReorderModels(values, false)
	for i := len(values) - 1; i >= 0; i-- {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			excutedTx := tx.Exec("DROP TABLE IF EXISTS ?", m.CurrentTable(stmt))

			if excutedTx.Error != nil {
				return excutedTx.Error
			}
			sqlList = append(sqlList, loadStmtSQL(excutedTx, stmt.Table))
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return sqlList, nil
}
