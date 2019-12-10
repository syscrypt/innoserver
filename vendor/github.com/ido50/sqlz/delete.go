package sqlz

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
)

// DeleteStmt represents a DELETE statement
type DeleteStmt struct {
	*Statment
	Table       string
	Conditions  []WhereCondition
	UsingTables []string
	Return      []string
	execer      Ext
}

// DeleteFrom creates a new DeleteStmt object for the
// provided table
func (db *DB) DeleteFrom(table string) *DeleteStmt {
	return &DeleteStmt{
		Table:  table,
		execer: db.DB,
		Statment: &Statment{db.ErrHandlers},
	}
}

// DeleteFrom creates a new DeleteStmt object for the
// provided table
func (tx *Tx) DeleteFrom(table string) *DeleteStmt {
	return &DeleteStmt{
		Table:  table,
		execer: tx.Tx,
		Statment: &Statment{tx.ErrHandlers},
	}
}

// Using adds a USING clause for joining in a delete statement
func (stmt *DeleteStmt) Using(tables ...string) *DeleteStmt {
	stmt.UsingTables = append(stmt.UsingTables, tables...)
	return stmt
}

// Where creates one or more WHERE conditions for the DELETE statement.
// If multiple conditions are passed, they are considered AND conditions.
func (stmt *DeleteStmt) Where(conds ...WhereCondition) *DeleteStmt {
	stmt.Conditions = append(stmt.Conditions, conds...)
	return stmt
}

// Returning sets a RETURNING clause to receive values back from the
// database once executing the DELETE statement. Note that GetRow or
// GetAll must be used to execute the query rather than Exec to get
// back the values.
func (stmt *DeleteStmt) Returning(cols ...string) *DeleteStmt {
	stmt.Return = append(stmt.Return, cols...)
	return stmt
}

// ToSQL generates the DELETE statement's SQL and returns a list of
// bindings. It is used internally by Exec, but is exported if you
// wish to use it directly.
func (stmt *DeleteStmt) ToSQL(rebind bool) (asSQL string, bindings []interface{}) {
	var clauses = []string{"DELETE FROM " + stmt.Table}

	if len(stmt.UsingTables) > 0 {
		clauses = append(clauses, "USING "+strings.Join(stmt.UsingTables, ", "))
	}

	if len(stmt.Conditions) > 0 {
		whereClause, whereBindings := parseConditions(stmt.Conditions)
		bindings = append(bindings, whereBindings...)
		clauses = append(clauses, "WHERE "+whereClause)
	}

	if len(stmt.Return) > 0 {
		clauses = append(clauses, "RETURNING "+strings.Join(stmt.Return, ", "))
	}

	asSQL = strings.Join(clauses, " ")

	if rebind {
		if db, ok := stmt.execer.(*sqlx.DB); ok {
			asSQL = db.Rebind(asSQL)
		} else if tx, ok := stmt.execer.(*sqlx.Tx); ok {
			asSQL = tx.Rebind(asSQL)
		}
	}

	return asSQL, bindings
}

// Exec executes the DELETE statement, returning the standard
// sql.Result struct and an error if the query failed.
func (stmt *DeleteStmt) Exec() (res sql.Result, err error) {
	asSQL, bindings := stmt.ToSQL(true)
	res, err = stmt.execer.Exec(asSQL, bindings...)
	stmt.HandlerError(err)
	return res, err
}

// ExecContext executes the DELETE statement, returning the standard
// sql.Result struct and an error if the query failed.
func (stmt *DeleteStmt) ExecContext(ctx context.Context) (res sql.Result, err error) {
	asSQL, bindings := stmt.ToSQL(true)
	res, err = stmt.execer.ExecContext(ctx, asSQL, bindings...)
	stmt.HandlerError(err)
	return res, err
}

// GetRow executes a DELETE statement with a RETURNING clause
// expected to return one row, and loads the result into
// the provided variable (which may be a simple variable if
// only one column is returned, or a struct if multiple columns
// are returned)
func (stmt *DeleteStmt) GetRow(into interface{}) error {
	asSQL, bindings := stmt.ToSQL(true)
	err := sqlx.Get(stmt.execer, into, asSQL, bindings...)
	stmt.HandlerError(err)
	return err
}

// GetRowContext executes a DELETE statement with a RETURNING clause
// expected to return one row, and loads the result into
// the provided variable (which may be a simple variable if
// only one column is returned, or a struct if multiple columns
// are returned)
func (stmt *DeleteStmt) GetRowContext(ctx context.Context, into interface{}) error {
	asSQL, bindings := stmt.ToSQL(true)
	err := sqlx.GetContext(ctx, stmt.execer, into, asSQL, bindings...)
	stmt.HandlerError(err)
	return err
}

// GetAll executes a DELETE statement with a RETURNING clause
// expected to return multiple rows, and loads the result into
// the provided slice variable
func (stmt *DeleteStmt) GetAll(into interface{}) error {
	asSQL, bindings := stmt.ToSQL(true)
	err := sqlx.Select(stmt.execer, into, asSQL, bindings...)
	stmt.HandlerError(err)
	return err
}

// GetAllContext executes a DELETE statement with a RETURNING clause
// expected to return multiple rows, and loads the result into
// the provided slice variable
func (stmt *DeleteStmt) GetAllContext(ctx context.Context, into interface{}) error {
	asSQL, bindings := stmt.ToSQL(true)
	err := sqlx.SelectContext(ctx, stmt.execer, into, asSQL, bindings...)
	stmt.HandlerError(err)
	return err
}
