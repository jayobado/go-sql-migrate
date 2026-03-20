package migrate

import (
	"fmt"
	"strings"
	"log/slog"

	"database/sql"
	"github.com/jmoiron/sqlx"
)

type ColumnInfo struct {
	Name      string
	Type      string
	Cid       *int
	NotNull   *int
	DfltValue *sql.NullString
	Pk        *int
}

func sqlQuery(tableName string, dialect SQLDialect) string {
	tableName = strings.ReplaceAll(tableName, `"`, "")
	switch dialect {
	case PostgreSQL:
		return fmt.Sprintf(`SELECT column_name AS name, data_type AS type FROM information_schema.columns WHERE table_name = '%s'`, tableName)
	case MySQL:
		return fmt.Sprintf(`SELECT column_name AS name, data_type AS type FROM information_schema.columns WHERE table_name = '%s'`, tableName)
	case SQLite:
		return fmt.Sprintf(`PRAGMA table_info("%s")`, tableName)
	default:
		return ""
	}
}

func GetTableColumns(db *sqlx.DB, tableName string, dialect SQLDialect) (map[string]string, error) {
	var cols []ColumnInfo

	query := sqlQuery(tableName, dialect)
	if query == "" {
		return nil, fmt.Errorf("unsupported SQL dialect: %s", dialect)
	}
	slog.Info("Executing query to get columns for", "table", tableName, "query", query)

	if err := db.Select(&cols, query); err != nil {
		return nil, fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	colsMap := make(map[string]string)
	for _, col := range cols {
		colName := strings.ToLower(col.Name)
		colType := strings.ToUpper(col.Type)
		if col.Cid != nil {
			slog.Info("Column has Cid", "name", colName, "key", *col.Cid)
		}
		if col.NotNull != nil {
			slog.Info("Column has NotNull", "name", colName, "key", *col.NotNull)
		}
		if col.DfltValue != nil && col.DfltValue.Valid {
			slog.Info("Column has Default Value", 
				"name", colName, 
				slog.String("value", col.DfltValue.String),
			)
		}
		if col.Pk != nil {
			slog.Info("Primary Key:", "name", colName, "key", *col.Pk)
		}
		colsMap[colName] = colType
		slog.Info("", "name", colName, "type", colType)
	}
	slog.Debug("colums fetched", 
		"table", tableName,
		"map", colsMap,
	)
	return colsMap, nil
}
