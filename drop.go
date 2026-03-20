package migrate

import (
	"fmt"
)

func DropTableSQL(schema Schema) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS %s;", schema.TableName())
}

func DropIndexSQL(tableName, column string, dialect SQLDialect) string {
	switch dialect {
	case PostgreSQL:
		return fmt.Sprintf("DROP INDEX IF EXISTS idx_%s_%s;", tableName, column)
	case MySQL:
		return fmt.Sprintf("DROP INDEX idx_%s_%s ON %s;", tableName, column, tableName)
	case SQLite:
		return fmt.Sprintf("DROP INDEX IF EXISTS idx_%s_%s;", tableName, column)
	default:
		return ""
	}
}

func DropConstraintSQL(tableName, constraintName string, dialect SQLDialect) string {
	switch dialect {
	case PostgreSQL:
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;", tableName, constraintName)
	case MySQL:
		return fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s;", tableName, constraintName)
	case SQLite:
		return "-- SQLite does not support DROP CONSTRAINT; requires full table rebuild"
	default:
		return ""
	}
}
