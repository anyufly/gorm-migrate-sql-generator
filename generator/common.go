package generator

import (
	"fmt"
	"github.com/anyufly/migrate-sql-result"
	"gorm.io/gorm"
)

type InvalidDialector struct {
	Tx *gorm.DB
}

func (i *InvalidDialector) Error() string {
	return fmt.Sprintf("unexcepted dialector: %s", i.Tx.Dialector.Name())
}

type MigrateSQLGenerator interface {
	Generate(values ...interface{}) (*result.MigrateSQLResult, error)
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

func GenMigrateSQL(tx *gorm.DB, values ...interface{}) (*result.MigrateSQLResult, error) {
	generator, err := loadGeneratorFromTx(tx)
	if err != nil {
		return nil, err
	}

	return generator.Generate(values...)
}
