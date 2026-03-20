# go-sql-migrate

A lightweight, reflection-based SQL schema migration library for Go. Supports PostgreSQL, MySQL, and SQLite.

## Installation
```bash
go get github.com/jayobado/go-sql-migrate
```

## Dependencies

This module uses [sqlx](https://github.com/jmoiron/sqlx) for database interaction. A `*sqlx.DB` instance is required to use `Migrate` and `GenerateSchemaDiffs`.
```bash
go get github.com/jmoiron/sqlx
```

You will also need a driver for your database:
```bash
# PostgreSQL
go get github.com/jackc/pgx/v5/stdlib

# MySQL
go get github.com/go-sql-driver/mysql

# SQLite
go get github.com/mattn/go-sqlite3
```

Register the driver with sqlx before passing the `*sqlx.DB` to this module:
```go
// PostgreSQL example
import (
    "github.com/jmoiron/sqlx"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sqlx.Open("pgx", "postgres://user:pass@localhost:5432/mydb")
if err != nil {
    log.Fatal(err)
}
```

## Usage

### Define a schema

Implement the `Schema` interface by adding a `TableName()` method to your struct, then annotate fields with struct tags:
```go
type User struct {
    ID        uuid.UUID `db:"id"        sql:"UUID"         primary:"true"`
    Email     string    `db:"email"     unique:"true"`
    Name      string    `db:"name"`
    Active    bool      `db:"active"    default:"true"`
    CreatedAt time.Time `db:"created_at"`
}

func (u User) TableName() string { return "users" }
```

### Run a migration
```go
import migrate "github.com/jayobado/go-sql-migrate"

schemas := []migrate.Schema{
    User{},
}

err := migrate.Migrate(ctx, db, migrate.ActionCreate, migrate.PostgreSQL, schemas)
if err != nil {
    log.Fatal(err)
}
```

Pass multiple schemas in one call:
```go
schemas := []migrate.Schema{
    User{},
    Post{},
    Comment{},
}

err := migrate.Migrate(ctx, db, migrate.ActionCreate, migrate.PostgreSQL, schemas)
if err != nil {
    log.Fatal(err)
}
```

## Actions

| Action | Constant | Description |
|--------|----------|-------------|
| Create | `migrate.ActionCreate` | Runs `CREATE TABLE IF NOT EXISTS` |
| Drop | `migrate.ActionDrop` | Runs `DROP TABLE IF EXISTS` |
| Alter | `migrate.ActionAlter` | Adds missing columns and updates changed column types |

## Dialects

| Dialect | Constant |
|---------|----------|
| PostgreSQL | `migrate.PostgreSQL` |
| MySQL | `migrate.MySQL` |
| SQLite | `migrate.SQLite` |

## Struct tags

| Tag | Description | Example |
|-----|-------------|---------|
| `db` | Column name. Omit or set to `-` to skip the field | `db:"email"` |
| `sql` | Override the SQL type | `sql:"VARCHAR(100)"` |
| `primary` | Mark as primary key | `primary:"true"` |
| `unique` | Add UNIQUE constraint | `unique:"true"` |
| `nullable` | Allow NULL values | `nullable:"true"` |
| `default` | Set a DEFAULT value | `default:"true"` |
| `fk` | Add a FOREIGN KEY reference | `fk:"other_table(id)"` |

## Go to SQL type mapping

| Go type | PostgreSQL | MySQL | SQLite |
|---------|-----------|-------|--------|
| `string` | `VARCHAR(255)` | `VARCHAR(255)` | `TEXT` |
| `int`, `int32` | `INTEGER` | `INTEGER` | `INTEGER` |
| `int64` | `INTEGER` | `BIGINT` | `INTEGER` |
| `float32`, `float64` | `NUMERIC` | `NUMERIC` | `NUMERIC` |
| `bool` | `BOOLEAN` | `TINYINT(1)` | `BOOLEAN` |
| `time.Time` | `TIMESTAMPTZ` | `DATETIME` | `DATETIME` |
| `uuid.UUID` | `UUID` | `CHAR(36)` | `TEXT` |
| `json.RawMessage` | `JSONB` | `JSON` | `TEXT` |

Override any mapping with the `sql` tag.

## Alter behaviour

`ActionAlter` compares your struct definition against the live table and generates statements for:
- Columns present in the struct but missing from the table → `ADD COLUMN`
- Columns where the Go-mapped type differs from the actual column type → `ALTER COLUMN ... TYPE`

Columns present in the table but absent from the struct are left untouched.

> **Note:** `ALTER COLUMN ... TYPE` can fail if existing data is incompatible with the new type. Review generated statements before running against production data.

## Generating SQL without executing

Use the underlying functions directly to inspect generated SQL before running it:
```go
// Create
sql, err := migrate.GenerateCreateTableSQL(User{}, migrate.PostgreSQL)

// Alter / diff
stmts, err := migrate.GenerateSchemaDiffs(db, User{}, migrate.PostgreSQL)
for _, s := range stmts {
    fmt.Println(s)
}

// Drop
sql := migrate.DropTableSQL(User{})
```

## License

[MIT](LICENSE)