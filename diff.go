package migrate

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
)

func GenerateSchemaDiffs(db *sqlx.DB, schema any, dialect SQLDialect) ([]string, error) {
	tn, ok := schema.(Schema)
	if !ok {
		return nil, fmt.Errorf("schema does not implement TableName() string")
	}
	tableName := tn.TableName()

	existing, err := GetTableColumns(db, tableName, dialect)
	if err != nil {
		return nil, err
	}

	typ := reflect.TypeOf(schema)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	var alters []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		colName := field.Tag.Get("db")
		if colName == "" || colName == "-" {
			continue
		}
		colKey := strings.ToLower(colName)

		sqlType := field.Tag.Get("sql")
		if sqlType == "" {
			sqlType = Map(field.Type, dialect)
		}
		upperType := strings.ToUpper(sqlType)

		actualType, exists := existing[colKey]

		switch {
		case !exists:
			slog.Debug("column missing, generating ADD COLUMN", "column", colKey, "table", tableName)
			stmt := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, colName, sqlType)

			if field.Tag.Get("nullable") != "true" && !strings.Contains(sqlType, "JSON") {
				stmt += " NOT NULL"
			}
			if def := field.Tag.Get("default"); def != "" {
				if strings.Contains(sqlType, "CHAR") || strings.Contains(sqlType, "TEXT") {
					stmt += fmt.Sprintf(" DEFAULT '%s'", def)
				} else {
					stmt += fmt.Sprintf(" DEFAULT %s", def)
				}
			}
			if field.Tag.Get("unique") == "true" {
				stmt += " UNIQUE"
			}
			if field.Tag.Get("primary") == "true" {
				stmt += " PRIMARY KEY"
			}
			if fk := field.Tag.Get("fk"); fk != "" {
				stmt += fmt.Sprintf(" REFERENCES %s", fk)
			}

			alters = append(alters, stmt+";")

		case !strings.EqualFold(actualType, upperType):
			slog.Debug("column type mismatch, generating ALTER TYPE",
				"column", colKey,
				"table", tableName,
				"expected", upperType,
				"actual", actualType,
			)
			alters = append(alters,
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;", tableName, colName, sqlType),
			)

		default:
			slog.Debug("column matches, skipping", "column", colKey, "table", tableName)
		}
	}

	return alters, nil
}