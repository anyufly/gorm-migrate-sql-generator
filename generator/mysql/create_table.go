package mysql

import (
	"fmt"
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"strings"
)

func (sqlGenerator *migrateSQLGenerator) CreateTable(execTx *gorm.DB, values ...interface{}) ([]*generator.SQLForTable, error) {
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	sqlList := make([]*generator.SQLForTable, 0, 10)

	for _, value := range m.ReorderModels(values, false) {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) (err error) {
			var (
				createTableSQL          = "CREATE TABLE IF NOT EXISTS ? ("
				values                  = []interface{}{m.CurrentTable(stmt)}
				hasPrimaryKeyInDataType bool
			)

			for _, dbName := range stmt.Schema.DBNames {
				field := stmt.Schema.FieldsByDBName[dbName]
				if !field.IgnoreMigration {
					createTableSQL += "? ?"
					hasPrimaryKeyInDataType = hasPrimaryKeyInDataType || strings.Contains(strings.ToUpper(string(field.DataType)), "PRIMARY KEY")
					values = append(values, clause.Column{Name: dbName}, m.DB.Migrator().FullDataTypeOf(field))
					createTableSQL += ","
				}
			}

			if !hasPrimaryKeyInDataType && len(stmt.Schema.PrimaryFields) > 0 {
				createTableSQL += "PRIMARY KEY ?,"
				primaryKeys := make([]interface{}, 0, len(stmt.Schema.PrimaryFields))
				for _, field := range stmt.Schema.PrimaryFields {
					primaryKeys = append(primaryKeys, clause.Column{Name: field.DBName})
				}

				values = append(values, primaryKeys)
			}

			var idxAfterCreateTable []string

			for _, idx := range stmt.Schema.ParseIndexes() {
				if m.CreateIndexAfterCreateTable {
					idxAfterCreateTable = append(idxAfterCreateTable, idx.Name)
				} else {
					if idx.Class != "" {
						createTableSQL += idx.Class + " "
					}
					createTableSQL += "INDEX ? ?"

					if idx.Comment != "" {
						createTableSQL += fmt.Sprintf(" COMMENT '%s'", idx.Comment)
					}

					if idx.Option != "" {
						createTableSQL += " " + idx.Option
					}

					createTableSQL += ","
					values = append(values, clause.Column{Name: idx.Name}, tx.Migrator().(migrator.BuildIndexOptionsInterface).BuildIndexOptions(idx.Fields, stmt))
				}
			}

			if !m.DB.DisableForeignKeyConstraintWhenMigrating && !m.DB.IgnoreRelationshipsWhenMigrating {
				for _, rel := range stmt.Schema.Relationships.Relations {
					if rel.Field.IgnoreMigration {
						continue
					}
					if constraint := rel.ParseConstraint(); constraint != nil {
						if constraint.Schema == stmt.Schema {
							sql, vars := buildConstraint(constraint)
							createTableSQL += sql + ","
							values = append(values, vars...)
						}
					}
				}
			}

			for _, chk := range stmt.Schema.ParseCheckConstraints() {
				createTableSQL += "CONSTRAINT ? CHECK (?),"
				values = append(values, clause.Column{Name: chk.Name}, clause.Expr{SQL: chk.Constraint})
			}

			createTableSQL = strings.TrimSuffix(createTableSQL, ",")

			createTableSQL += ")"

			if tableOption, ok := m.DB.Get("gorm:table_options"); ok {
				createTableSQL += fmt.Sprint(tableOption)
			}

			executedTx := tx.Exec(createTableSQL, values...)

			if executedTx.Error != nil {
				return executedTx.Error
			}

			sqlList = append(sqlList, loadStmtSQL(executedTx, stmt.Table))

			if m.CreateIndexAfterCreateTable {
				for _, idx := range idxAfterCreateTable {
					sql, err := sqlGenerator.CreateIndex(tx, value, idx)
					if err != nil {
						return err
					}
					sqlList = append(sqlList, sql)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	return sqlList, nil
}
