package migrate

import (
	"fmt"
	"strings"
	"reflect"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

func GenerateAlterTableAddColumns(db *sqlx.DB, schema any, dialect SQLDialect) ([]string, error) {
	tn, ok := schema.(Schema)
	if !ok {
		return nil, fmt.Errorf("model does not implement TableName() string")
	}
	tableName := tn.TableName()

	schemaCols, err := GetTableColumns(db, tableName, dialect)
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

		if _, exists := schemaCols[colKey]; exists {
			slog.Debug("Skipping existing column", "key", colKey)
			continue
		} else {
			slog.Debug("→ NOT FOUND in DB: column — will be added", "key", colKey)
		}

		// Use the original colName for SQL generation
		slog.Debug("Adding column", "name", colName)

		sqlType := field.Tag.Get("sql")
		if sqlType == "" {
			sqlType = Map(field.Type, dialect)
		}

		line := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, colName, sqlType)

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

		if field.Tag.Get("unique") == "true" {
			line += " UNIQUE"
		}
		if field.Tag.Get("primary") == "true" {
			line += " PRIMARY KEY"
		}
		if fk := field.Tag.Get("fk"); fk != "" {
			line += fmt.Sprintf(" REFERENCES %s", fk)
		}

		alters = append(alters, line+";")
	}

	return alters, nil
}

func GenerateColumnTypeDiffs(db *sqlx.DB, schema any, dialect SQLDialect) ([]string, error) {
	tn, ok := schema.(Schema)
	if !ok {
		return nil, fmt.Errorf("schema does not implement TableName() string")
	}
	tableName := tn.TableName()
	schemaCols, err := GetTableColumns(db, tableName, dialect)
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

		sqlType := field.Tag.Get("sql")
		if sqlType == "" {
			sqlType = Map(field.Type, dialect)
		}

		if actualType, ok := schemaCols[colName]; ok && !strings.EqualFold(actualType, strings.ToUpper(sqlType)) {
			alters = append(alters, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;", tableName, colName, sqlType))
		}
	}
	return alters, nil
}


func GenerateSchemaDiffs(db *sqlx.DB, schema any, dialect SQLDialect) ([]string, error) {
	tn, ok := schema.(Schema)
	if !ok {
		return nil, fmt.Errorf("model does not implement TableName() string")
	}
	tableName := tn.TableName()

	schemaCols, err := GetTableColumns(db, tableName, dialect)
	if err != nil {
		return nil, err
	}

	// Normalize schema column keys to lowercase
	normalizedCols := map[string]string{}
	for name, typ := range schemaCols {
		normalizedCols[strings.ToLower(name)] = typ
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

		upperSQLType := strings.ToUpper(sqlType)
		actualType, exists := normalizedCols[colKey]

		switch {
		case !exists:
			slog.Info("→ Column not found, generating ADD COLUMN", slog.String("key", colKey))
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

		case !strings.EqualFold(actualType, upperSQLType):
			slog.Debug("→ Column %s exists but differs (expected %s, actual %s), generating ALTER TYPE", 
				"col_key", colKey, 
				"expected", upperSQLType, 
				"actual", actualType,
			)
			stmt := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;", tableName, colName, sqlType)
			alters = append(alters, stmt)

		default:
			slog.Debug("→ Column %s matches existing schema, skipping", "col_key", colKey)
		}
	}

	return alters, nil
}
