package mysql

import (
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
)

func init() {
	generator.Register("mysql", MigrateSQLGenerator)
}

type migrateSQLGenerator struct {
	tx *gorm.DB
}

func MigrateSQLGenerator(tx *gorm.DB) generator.MigrateSQLGenerator {
	return &migrateSQLGenerator{
		tx: tx,
	}
}

func (sqlGenerator *migrateSQLGenerator) Generate(values ...interface{}) (*result.MigrateSQLResult, error) {
	m, err := loadMigrator(sqlGenerator.tx)

	if err != nil {
		return nil, err
	}

	migrateResult := result.NewMigrateSQLResult()

	for _, value := range m.ReorderModels(values, true) {
		queryTx := m.DB.Session(&gorm.Session{
			DryRun: false,
		})

		execTx := m.DB.Session(&gorm.Session{
			DryRun: true,
		})

		if !queryTx.Migrator().HasTable(value) {
			up, err := sqlGenerator.CreateTable(execTx, value)
			if err != nil {
				return nil, err
			}
			migrateResult.AppendUp(up...)

			down, err := sqlGenerator.DropTable(execTx, value)
			if err != nil {
				return nil, err
			}
			migrateResult.AppendDown(down...)
		} else {
			if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
				columnTypes, err := queryTx.Migrator().ColumnTypes(value)
				if err != nil {
					return err
				}
				var (
					parseIndexes          = stmt.Schema.ParseIndexes()
					parseCheckConstraints = stmt.Schema.ParseCheckConstraints()
				)
				for _, dbName := range stmt.Schema.DBNames {
					var foundColumn gorm.ColumnType

					for _, columnType := range columnTypes {
						if columnType.Name() == dbName {
							foundColumn = columnType
							break
						}
					}

					if foundColumn == nil {
						up, err := sqlGenerator.AddColumn(execTx, value, dbName)
						if err != nil {
							return err
						}
						if up != nil {
							migrateResult.AppendUp(up)
						}

						down, err := sqlGenerator.DropColumn(execTx, value, dbName)
						if err != nil {
							return err
						}
						if down != nil {
							migrateResult.AppendDown(down)
						}

					} else {
						// found, smartly migrate
						field := stmt.Schema.FieldsByDBName[dbName]

						up, err := sqlGenerator.MigrateColumn(execTx, value, field, foundColumn)
						if err != nil {
							return err
						}

						if up != nil {
							migrateResult.AppendUp(up)
						}

						down, err := sqlGenerator.MigrateColumnRecover(execTx, value, field, foundColumn)
						if err != nil {
							return err
						}

						if down != nil {
							migrateResult.AppendDown(down)
						}
					}
				}

				if !m.DB.DisableForeignKeyConstraintWhenMigrating && !m.DB.IgnoreRelationshipsWhenMigrating {
					for _, rel := range stmt.Schema.Relationships.Relations {
						if rel.Field.IgnoreMigration {
							continue
						}
						if constraint := rel.ParseConstraint(); constraint != nil &&
							constraint.Schema == stmt.Schema && !queryTx.Migrator().HasConstraint(value, constraint.Name) {

							up, err := sqlGenerator.CreateConstraint(execTx, value, constraint.Name)
							if err != nil {
								return err
							}

							if up != nil {
								migrateResult.AppendUp(up)
							}

							down, err := sqlGenerator.DropConstraint(execTx, value, constraint.Name)
							if err != nil {
								return err
							}

							if down != nil {
								migrateResult.AppendDown(down)
							}

						}
					}
				}

				for _, chk := range parseCheckConstraints {
					if !queryTx.Migrator().HasConstraint(value, chk.Name) {
						up, err := sqlGenerator.CreateConstraint(execTx, value, chk.Name)
						if err != nil {
							return err
						}

						if up != nil {
							migrateResult.AppendUp(up)
						}

						down, err := sqlGenerator.DropConstraint(execTx, value, chk.Name)
						if err != nil {
							return err
						}

						if down != nil {
							migrateResult.AppendDown(down)
						}
					}
				}

				for _, idx := range parseIndexes {
					if !queryTx.Migrator().HasIndex(value, idx.Name) {

						up, err := sqlGenerator.CreateIndex(execTx, value, idx.Name)
						if err != nil {
							return err
						}
						if up != nil {
							migrateResult.AppendUp(up)
						}

						down, err := sqlGenerator.DropIndex(execTx, value, idx.Name)
						if err != nil {
							return err
						}

						if down != nil {
							migrateResult.AppendDown(down)
						}
					}
				}

				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	return migrateResult, nil
}
