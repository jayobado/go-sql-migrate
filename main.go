package migrate

import (
	"fmt"
	"errors"
	"context"
	
	"github.com/jmoiron/sqlx"
)

func process(
	ctx context.Context,
	db *sqlx.DB, 
	action string,
	schema Schema,
	dialect *SQLDialect,
) error {

	switch action {
	case "create":
		query, err := GenerateCreateTableSQL(schema, PostgreSQL)
		if err != nil {
			return err
		}
		if _, err = db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error creating table %s: %v", schema.TableName(), err)
		}
	case "drop":
		query := DropTableSQL(schema)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error dropping table %s: %v", schema.TableName(), err)
		}
	case "alter":
		if dialect == nil {
			return errors.New("dialect must be indicated to run alter command")
		}

		queries, err := GenerateAlterTableAddColumns(db, schema, *dialect)
		if err != nil {
			return err
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("error altering table %s: %v", schema.TableName(), err)
			}
		}
	default:
		return errors.New("action entered is incorrect")
	}
	return nil
}

func Migrate(ctx context.Context, db *sqlx.DB, action string, schemas []Schema, dialect *SQLDialect) error {
	for _, schema := range schemas {
		if err := process(ctx, db, action, schema, dialect); err != nil {
			return err
		}
	}
	return nil
}