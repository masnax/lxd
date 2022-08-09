package query

import (
	"database/sql"
	"fmt"
	"strings"
)

// SelectObjects executes a statement which must yield rows with a specific
// columns schema. It invokes the given Dest hook for each yielded row.
func SelectObjects(stmt *sql.Stmt, dest Dest, args ...any) error {
	rows, err := stmt.Query(args...)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for i := 0; rows.Next(); i++ {
		err := rows.Scan(dest(i)...)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

// QueryObjects executes a statement from a query string which must yield rows
// with a specific columns schema. It invokes the given Dest hook for each yielded row.
func QueryObjects(tx *sql.Tx, stmt string, dest Dest, args ...any) error {
	rows, err := tx.Query(stmt, args...)
	if err != nil {
		return err
	}

	defer func() { _ = rows.Close() }()

	for i := 0; rows.Next(); i++ {
		err := rows.Scan(dest(i)...)
		if err != nil {
			return err
		}
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}

// Dest is a function that is expected to return the objects to pass to the
// 'dest' argument of sql.Rows.Scan(). It is invoked by SelectObjects once per
// yielded row, and it will be passed the index of the row being scanned.
type Dest func(i int) []any

// UpsertObject inserts or replaces a new row with the given column values, to
// the given table using columns order. For example:
//
// UpsertObject(tx, "cars", []string{"id", "brand"}, []any{1, "ferrari"})
//
// The number of elements in 'columns' must match the one in 'values'.
func UpsertObject(tx *sql.Tx, table string, columns []string, values []any) (int64, error) {
	n := len(columns)
	if n == 0 {
		return -1, fmt.Errorf("columns length is zero")
	}

	if n != len(values) {
		return -1, fmt.Errorf("columns length does not match values length")
	}

	stmt := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES %s",
		table, strings.Join(columns, ", "), Params(n))
	result, err := tx.Exec(stmt, values...)
	if err != nil {
		return -1, fmt.Errorf("insert or replaced row: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("get last inserted ID: %w", err)
	}

	return id, nil
}

// DeleteObject removes the row identified by the given ID. The given table
// must have a primary key column called 'id'.
//
// It returns a flag indicating if a matching row was actually found and
// deleted or not.
func DeleteObject(tx *sql.Tx, table string, id int64) (bool, error) {
	stmt := fmt.Sprintf("DELETE FROM %s WHERE id=?", table)
	result, err := tx.Exec(stmt, id)
	if err != nil {
		return false, err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if n > 1 {
		return true, fmt.Errorf("more than one row was deleted")
	}

	return n == 1, nil
}
