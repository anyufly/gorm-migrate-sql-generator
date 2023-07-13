package mysql

import (
	"fmt"
	result "github.com/anyufly/migrate-sql-result"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"regexp"
	"strings"
)

var regFullDataType = regexp.MustCompile(`\D*(\d+)\D?`)

func shouldAlterColumn(m mysql.Migrator, field *schema.Field, columnType gorm.ColumnType) bool {
	fullDataType := strings.TrimSpace(strings.ToLower(m.DB.Migrator().FullDataTypeOf(field).SQL))
	realDataType := strings.ToLower(columnType.DatabaseTypeName())

	var (
		alterColumn bool
		isSameType  = fullDataType == realDataType
	)

	if !field.PrimaryKey {
		// check type
		if !strings.HasPrefix(fullDataType, realDataType) {
			// check type aliases
			aliases := m.DB.Migrator().GetTypeAliases(realDataType)
			for _, alias := range aliases {
				if strings.HasPrefix(fullDataType, alias) {
					isSameType = true
					break
				}
			}

			if !isSameType {
				alterColumn = true
			}
		}
	}

	if !isSameType {
		// check size
		if length, ok := columnType.Length(); length != int64(field.Size) {
			if length > 0 && field.Size > 0 {
				alterColumn = true
			} else {
				// has size in data type and not equal
				// Since the following code is frequently called in the for loop, reg optimization is needed here
				matches2 := regFullDataType.FindAllStringSubmatch(fullDataType, -1)
				if !field.PrimaryKey &&
					(len(matches2) == 1 && matches2[0][1] != fmt.Sprint(length) && ok) {
					alterColumn = true
				}
			}
		}

		// check precision
		if precision, _, ok := columnType.DecimalSize(); ok && int64(field.Precision) != precision {
			if regexp.MustCompile(fmt.Sprintf("[^0-9]%d[^0-9]", field.Precision)).MatchString(m.Migrator.DataTypeOf(field)) {
				alterColumn = true
			}
		}
	}

	// check nullable
	if nullable, ok := columnType.Nullable(); ok && nullable == field.NotNull {
		// not primary key & database is nullable
		if !field.PrimaryKey && nullable {
			alterColumn = true
		}
	}

	// check unique
	if unique, ok := columnType.Unique(); ok && unique != field.Unique {
		// not primary key
		if !field.PrimaryKey {
			alterColumn = true
		}
	}

	// check default value
	if !field.PrimaryKey {
		currentDefaultNotNull := field.HasDefaultValue && (field.DefaultValueInterface != nil || !strings.EqualFold(field.DefaultValue, "NULL"))
		dv, dvNotNull := columnType.DefaultValue()
		if dvNotNull && !currentDefaultNotNull {
			// default value -> null
			alterColumn = true
		} else if !dvNotNull && currentDefaultNotNull {
			// null -> default value
			alterColumn = true
		} else if (field.GORMDataType != schema.Time && dv != field.DefaultValue) ||
			(field.GORMDataType == schema.Time && !strings.EqualFold(strings.TrimSuffix(dv, "()"), strings.TrimSuffix(field.DefaultValue, "()"))) {
			// default value not equal
			// not both null
			if currentDefaultNotNull || dvNotNull {
				alterColumn = true
			}
		}
	}

	// check comment
	if comment, ok := columnType.Comment(); ok && comment != field.Comment {
		// not primary key
		if !field.PrimaryKey {
			alterColumn = true
		}
	}

	return alterColumn
}

func (sqlGenerator *migrateSQLGenerator) MigrateColumn(
	execTx *gorm.DB, value interface{}, field *schema.Field, columnType gorm.ColumnType) (*result.SQLForTable, error) {

	var sql *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if shouldAlterColumn(m, field, columnType) && !field.IgnoreMigration {
		return sqlGenerator.AlterColumn(execTx, value, field.DBName)
	}

	return sql, nil
}

func (sqlGenerator *migrateSQLGenerator) MigrateColumnRecover(
	execTx *gorm.DB, value interface{}, field *schema.Field, columnType gorm.ColumnType) (*result.SQLForTable, error) {

	var sql *result.SQLForTable
	m, err := loadMigratorWithDryRun(execTx)
	if err != nil {
		return nil, err
	}

	if shouldAlterColumn(m, field, columnType) && !field.IgnoreMigration {
		return sqlGenerator.RecoverAlter(execTx, value, columnType)
	}

	return sql, nil
}
