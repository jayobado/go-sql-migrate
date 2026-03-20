package migrate

import (
	"fmt"
	"reflect"
	"log/slog"
)

type SQLDialect string

const (
	PostgreSQL SQLDialect = "postgres"
	MySQL      SQLDialect = "mysql"
	SQLite     SQLDialect = "sqlite"
)

func (prop SQLDialect) Validate() error {
	switch prop {
	case PostgreSQL, MySQL, SQLite:
		return nil
	default:
		return fmt.Errorf("invalid sql dialect: %s", prop)
	}
}


func Map(t reflect.Type, dialect SQLDialect) string {
	slog.Info("Mapping type %s to SQL for dialect %s", t.Name(), dialect)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Name() {
	case "UUID":
		switch dialect {
		case PostgreSQL:
			return "UUID"
		case MySQL:
			return "CHAR(36)"
		case SQLite:
			return "TEXT"
		}
	case "string":
		switch dialect {
		case PostgreSQL, MySQL:
			return "VARCHAR(255)"
		case SQLite:
			return "TEXT"
		}
	case "float32", "float64":
		return "NUMERIC"
	case "int", "int32":
		return "INTEGER"
	case "int64":
		if dialect == MySQL {
			return "BIGINT"
		}
		return "INTEGER"
	case "bool":
		if dialect == MySQL {
			return "TINYINT(1)"
		}
		return "BOOLEAN"
	case "Time":
		switch dialect {
		case PostgreSQL:
			return "TIMESTAMPTZ"
		case MySQL, SQLite:
			return "DATETIME"
		}
	case "RawMessage":
		switch dialect {
		case PostgreSQL:
			return "JSONB"
		case MySQL:
			return "JSON"
		case SQLite:
			return "TEXT"
		}
	}
	return "TEXT"
}