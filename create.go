package migrate

import (
	"fmt"
	"strings"
	"reflect"
)


type Schema interface {
	TableName() string
}

func GenerateCreateTableSQL(schema any, dialect SQLDialect) (string, error) {
	tn, ok := schema.(Schema)
	if !ok {
		return "", fmt.Errorf("model does not implement TableName() string")
	}
	tableName := tn.TableName()
	typ := reflect.TypeOf(schema)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	var columns []string
	var foreignKeys []string

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
	if len(foreignKeys) > 0 {
		columns = append(columns, foreignKeys...)
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columns, ", ")), nil
}
