package migrate

import (
	"fmt"
	"reflect"
	"strings"
)

func GenerateCreateTableSQL(schema Schema, dialect SQLDialect) (string, error) {
	typ := reflect.TypeOf(schema)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	var columns, foreignKeys []string

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		colName := field.Tag.Get("db")
		if colName == "" || colName == "-" {
			continue
		}

		sqlType := field.Tag.Get("sql")
		if sqlType == "" {
			sqlType = Map(field.Type, dialect)
		}

		line := colName + " " + sqlType

		if field.Tag.Get("nullable") != "true" && !strings.Contains(sqlType, "JSON") {
			line += " NOT NULL"
		}
		if def := field.Tag.Get("default"); def != "" {
			if strings.Contains(sqlType, "CHAR") || strings.Contains(sqlType, "TEXT") {
				line += fmt.Sprintf(" DEFAULT '%s'", def)
			} else {
				line += fmt.Sprintf(" DEFAULT %s", def)
			}
		}
		if field.Tag.Get("primary") == "true" {
			line += " PRIMARY KEY"
		}
		if field.Tag.Get("unique") == "true" {
			line += " UNIQUE"
		}
		if fk := field.Tag.Get("fk"); fk != "" {
			foreignKeys = append(foreignKeys, fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s", colName, fk))
		}

		columns = append(columns, line)
	}

	columns = append(columns, foreignKeys...)

	return fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s);",
		schema.TableName(),
		strings.Join(columns, ", "),
	), nil
}