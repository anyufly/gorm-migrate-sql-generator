package mysql

import (
	"fmt"
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

func (sqlGenerator *migrateSQLGenerator) AlterColumn(execTx *gorm.DB, value interface{}, field string) (*generator.SQLForTable, error) {
	var sql *generator.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				fullDataType := m.FullDataTypeOf(field)
				if m.Dialector.DontSupportRenameColumnUnique {
					fullDataType.SQL = strings.Replace(fullDataType.SQL, " UNIQUE ", " ", 1)
				}

				executedTx := m.DB.Exec(
					"ALTER TABLE ? MODIFY COLUMN ? ?",
					clause.Table{Name: stmt.Table}, clause.Column{Name: field.DBName}, fullDataType,
				)

				if executedTx.Error != nil {
					return executedTx.Error
				}

				sql = loadStmtSQL(executedTx, stmt.Table)
				return nil
			}
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	}); err != nil {
		return nil, err
	}

	return sql, nil
}

func (sqlGenerator *migrateSQLGenerator) parseColumnType(m mysql.Migrator, columnType gorm.ColumnType) (expr clause.Expr) {
	if ct, ok := columnType.ColumnType(); ok {
		expr.SQL += ct
	}

	if nullable, ok := columnType.Nullable(); ok && !nullable {
		expr.SQL += " NOT NULL"
	}

	if unique, ok := columnType.Unique(); ok && unique {
		expr.SQL += " UNIQUE"
	}

	if value, ok := columnType.DefaultValue(); ok {
		expr.SQL += " DEFAULT " + m.Dialector.Explain("?", value)
	}

	if comment, ok := columnType.Comment(); ok {
		expr.SQL += " COMMENT " + m.Dialector.Explain("?", comment)
	}

	return
}

func (sqlGenerator *migrateSQLGenerator) RecoverAlter(execTx *gorm.DB, value interface{}, columnType gorm.ColumnType) (*generator.SQLForTable, error) {
	var sql *generator.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		executedTx := m.DB.Exec(
			"ALTER TABLE ? MODIFY COLUMN ? ?",
			m.CurrentTable(stmt), clause.Column{Name: columnType.Name()}, sqlGenerator.parseColumnType(m, columnType),
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
