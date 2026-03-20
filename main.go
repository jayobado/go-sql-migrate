package migrate

import (
	"fmt"
	"context"
	
	"github.com/jmoiron/sqlx"
)

type Schema interface {
	TableName() string
}

func process(
	ctx context.Context,
	db *sqlx.DB, 
	action Action,
	schema Schema,
	dialect SQLDialect,
) error {
	switch action {
	case ActionCreate:
		query, err := GenerateCreateTableSQL(schema, dialect)
		if err != nil {
			return err
		}
		if _, err = db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error creating table %s: %w", schema.TableName(), err)
		}
	case ActionDrop:
		query := DropTableSQL(schema)
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("error dropping table %s: %w", schema.TableName(), err)
		}
	case ActionAlter:
		queries, err := GenerateSchemaDiffs(db, schema, dialect)
		if err != nil {
			return err
		}

		for _, query := range queries {
			if _, err := db.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("error altering table %s: %w", schema.TableName(), err)
			}
		}
	}
	return nil
}

func Migrate(ctx context.Context, db *sqlx.DB, action Action, dialect SQLDialect, schemas []Schema) error {
	if err := action.Validate(); err != nil {
		return err
	}
	if err := dialect.Validate(); err != nil {
		return err
	}
	for _, schema := range schemas {
		if err := process(ctx, db, action, schema, dialect); err != nil {
			return err
		}
	}
	return nil
}