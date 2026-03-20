package migrate

import (
	"fmt"
	"reflect"
	"log/slog"
)

type SQLDialect string
type Action string

const (
	PostgreSQL SQLDialect = "postgres"
	MySQL      SQLDialect = "mysql"
	SQLite     SQLDialect = "sqlite"
)

const (
	ActionCreate Action = "create"
	ActionDrop   Action = "drop"
	ActionAlter  Action = "alter"
)

func (d SQLDialect) Validate() error {
	switch d {
	case PostgreSQL, MySQL, SQLite:
		return nil
	default:
		return fmt.Errorf("invalid sql dialect: %q", d)
	}
}

func (a Action) Validate() error {
	switch a {
	case ActionCreate, ActionDrop, ActionAlter:
		return nil
	default:
		return fmt.Errorf("invalid action: %q", a)
	}
}

func NewAction(value string) Action {
	switch value {
	case "create":
		return ActionCreate
	case "drop":
		return ActionDrop
	case "alter":
		return ActionAlter
	default:
		return ""
	}
}


func Map(t reflect.Type, dialect SQLDialect) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	var sqlType string

	switch t.Name() {
	case "UUID":
		switch dialect {
		case PostgreSQL:
			sqlType = "UUID"
		case MySQL:
			sqlType = "CHAR(36)"
		default:
			sqlType = "TEXT"
		}
	case "string":
		switch dialect {
		case PostgreSQL, MySQL:
			sqlType = "VARCHAR(255)"
		default:
			sqlType = "TEXT"
		}
	case "float32", "float64":
		sqlType = "NUMERIC"
	case "int", "int32":
		sqlType = "INTEGER"
	case "int64":
		if dialect == MySQL {
			sqlType = "BIGINT"
		} else {
			sqlType = "INTEGER"
		}
	case "bool":
		if dialect == MySQL {
			sqlType = "TINYINT(1)"
		} else {
			sqlType = "BOOLEAN"
		}
	case "Time":
		switch dialect {
		case PostgreSQL:
			sqlType = "TIMESTAMPTZ"
		default:
			sqlType = "DATETIME"
		}
	case "RawMessage":
		switch dialect {
		case PostgreSQL:
			sqlType = "JSONB"
		case MySQL:
			sqlType = "JSON"
		default:
			sqlType = "TEXT"
		}
	default:
		slog.Warn("unknown Go type, defaulting to TEXT", "type", t.Name(), "dialect", dialect)
		sqlType = "TEXT"
	}

	return sqlType
}