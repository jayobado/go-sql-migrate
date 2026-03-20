package migrate

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ColumnInfo struct {
	Name string         `db:"name"`
	Type string         `db:"type"`
	// SQLite-only fields — nil for Postgres/MySQL
	Cid       *int            `db:"cid"`
	NotNull   *int            `db:"notnull"`
	DfltValue *sql.NullString `db:"dflt_value"`
	Pk        *int            `db:"pk"`
}

func columnQuery(tableName string, dialect SQLDialect) (string, error) {
	tableName = strings.ReplaceAll(tableName, `"`, "")
	switch dialect {
	case PostgreSQL, MySQL:
		return fmt.Sprintf(
			`SELECT column_name AS name, data_type AS type FROM information_schema.columns WHERE table_name = '%s'`,
			tableName,
		), nil
	case SQLite:
		return fmt.Sprintf(`PRAGMA table_info("%s")`, tableName), nil
	default:
		return "", fmt.Errorf("unsupported dialect: %q", dialect)
	}
}

func GetTableColumns(db *sqlx.DB, tableName string, dialect SQLDialect) (map[string]string, error) {
	query, err := columnQuery(tableName, dialect)
	if err != nil {
		return nil, err
	}

	slog.Debug("fetching table columns", "table", tableName, "dialect", dialect)

	var cols []ColumnInfo
	if err := db.Select(&cols, query); err != nil {
		return nil, fmt.Errorf("get columns for %s: %w", tableName, err)
	}

	result := make(map[string]string, len(cols))
	for _, col := range cols {
		result[strings.ToLower(col.Name)] = strings.ToUpper(col.Type)
	}

	slog.Debug("columns fetched", "table", tableName, "count", len(result))
	return result, nil
}