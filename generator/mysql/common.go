package mysql

import (
	"github.com/anyufly/gorm-migrate-sql-generator/generator"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

func buildConstraint(constraint *schema.Constraint) (sql string, results []interface{}) {
	sql = "CONSTRAINT ? FOREIGN KEY ? REFERENCES ??"
	if constraint.OnDelete != "" {
		sql += " ON DELETE " + constraint.OnDelete
	}

	if constraint.OnUpdate != "" {
		sql += " ON UPDATE " + constraint.OnUpdate
	}

	var foreignKeys, references []interface{}
	for _, field := range constraint.ForeignKeys {
		foreignKeys = append(foreignKeys, clause.Column{Name: field.DBName})
	}

	for _, field := range constraint.References {
		references = append(references, clause.Column{Name: field.DBName})
	}
	results = append(results, clause.Table{Name: constraint.Name}, foreignKeys, clause.Table{Name: constraint.ReferenceSchema.Table}, references)
	return
}

func dryRun(tx *gorm.DB) *gorm.DB {
	if !tx.DryRun {
		tx = tx.Session(&gorm.Session{
			DryRun: true,
		})
	}

	return tx
}

func loadMigrator(tx *gorm.DB) (mysql.Migrator, error) {
	m, ok := tx.Migrator().(mysql.Migrator)

	if !ok {
		return mysql.Migrator{}, &generator.InvalidDialector{Tx: tx}
	}

	return m, nil
}

func loadMigratorWithDryRun(tx *gorm.DB) (mysql.Migrator, error) {
	dryRunTx := dryRun(tx)
	return loadMigrator(dryRunTx)
}

func loadStmtSQL(db *gorm.DB, table string) *generator.SQLForTable {
	sql, vars := db.Statement.SQL.String(), db.Statement.Vars
	explainedSQL := db.Dialector.Explain(sql, vars...)

	return generator.NewSQLForTable(table, explainedSQL)
}
