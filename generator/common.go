package generator

import (
	"fmt"
	"gorm.io/gorm"
)

type InvalidDialector struct {
	Tx *gorm.DB
}

func (i *InvalidDialector) Error() string {
	return fmt.Sprintf("unexcepted dialector: %s", i.Tx.Dialector.Name())
}

type SQLForTable struct {
	table string
	sql   string
}

func NewSQLForTable(table string, sql string) *SQLForTable {
	return &SQLForTable{
		table: table,
		sql:   sql,
	}
}

func (s *SQLForTable) Table() string {
	return s.table
}

func (s *SQLForTable) SQL() string {
	return s.sql
}

type SQLForTableList []*SQLForTable

func (sqlList SQLForTableList) ToMap() map[string][]string {
	sqlMap := make(map[string][]string)
	for _, sql := range sqlList {
		if sql != nil {
			if tableSQLList, ok := sqlMap[sql.Table()]; ok {
				tableSQLList = append(tableSQLList, sql.SQL())
				sqlMap[sql.Table()] = tableSQLList
			} else {
				sqlMap[sql.Table()] = []string{sql.SQL()}
			}
		}
	}
	return sqlMap
}

type MigrateSQLResult struct {
	up   SQLForTableList
	down SQLForTableList
}

func NewMigrateSQLResult() *MigrateSQLResult {
	return &MigrateSQLResult{
		up:   make([]*SQLForTable, 0, 10),
		down: make([]*SQLForTable, 0, 10),
	}
}

func (result *MigrateSQLResult) AppendUp(sql ...*SQLForTable) {
	result.up = append(result.up, sql...)
}

func (result *MigrateSQLResult) Up() map[string][]string {
	return result.up.ToMap()
}

func (result *MigrateSQLResult) AppendDown(sql ...*SQLForTable) {
	result.down = append(result.down, sql...)
}

func (result *MigrateSQLResult) Down() map[string][]string {
	return result.down.ToMap()
}

type MigrateSQLGenerator interface {
	Generate(values ...interface{}) (*MigrateSQLResult, error)
}

type MigrateSQLGeneratorMaker func(tx *gorm.DB) MigrateSQLGenerator

var generators = make(map[string]MigrateSQLGeneratorMaker)

func Register(name string, generatorMaker MigrateSQLGeneratorMaker) {
	generators[name] = generatorMaker
}

func loadGeneratorFromTx(tx *gorm.DB) (MigrateSQLGenerator, error) {
	if generatorMaker, ok := generators[tx.Dialector.Name()]; ok {
		return generatorMaker(tx), nil
	} else {
		return nil, &InvalidDialector{Tx: tx}
	}
}

func GenMigrateSQL(tx *gorm.DB, values ...interface{}) (*MigrateSQLResult, error) {
	generator, err := loadGeneratorFromTx(tx)
	if err != nil {
		return nil, err
	}

	return generator.Generate(values...)
}
