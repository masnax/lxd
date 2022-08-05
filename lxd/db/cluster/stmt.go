//go:build linux && cgo && !agent

package cluster

import (
	"database/sql"
	"fmt"
	"strings"
)

// RegisterStmt register a SQL statement.
//
// Registered statements will be prepared upfront and re-used, to speed up
// execution.
//
// Return a unique registration code.
//
// The additional parameter numFilters replaces the key '{num_filters}' in
// the query string with a list of parameters of length numFilters.
func RegisterStmt(sql string, numFilters ...int) int {
	count := 1
	if len(numFilters) == 1 {
		count = numFilters[0]
		sql = strings.Replace(sql, "{num_filters}", fmt.Sprintf("?%s", strings.Repeat(", ?", count-1)), -1)
	}

	code := len(stmts)
	stmts[code] = stmtConfig{query: sql, numFilters: count}
	return code
}

// PrepareStmts prepares all registered statements and returns an index from
// statement code to prepared statement object.
func PrepareStmts(db *sql.DB, skipErrors bool) (map[int]*sql.Stmt, error) {
	index := map[int]*sql.Stmt{}

	for code, sql := range stmts {
		stmt, err := db.Prepare(sql.query)
		if err != nil && !skipErrors {
			return nil, fmt.Errorf("%q: %w", sql, err)
		}

		index[code] = stmt
	}

	return index, nil
}

// stmtConfig represents information about a registered SQL statement.
type stmtConfig struct {
	query string

	numFilters int
}

var stmts = map[int]stmtConfig{} // Statement code to statement SQL text.

// PreparedStmts is a placeholder for transitioning to package-scoped transaction functions.
var PreparedStmts = map[int]*sql.Stmt{}

// Stmt prepares the in-memory prepared statement for the transaction.
func Stmt(tx *sql.Tx, code int) *sql.Stmt {
	stmt, ok := PreparedStmts[code]
	if !ok {
		panic(fmt.Sprintf("No prepared statement registered with code %d", code))
	}

	return tx.Stmt(stmt)
}

// NumFilters returns the number of filters that the statement expects. The default is 1.
func NumFilters(code int) int {
	stmt, ok := stmts[code]
	if !ok {
		panic(fmt.Sprintf("No prepared statement registered with code %d", code))
	}

	return stmt.numFilters
}

// prepare prepares a new statement from a SQL string.
func prepare(tx *sql.Tx, sql string) (*sql.Stmt, error) {
	stmt, err := tx.Prepare(sql)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare statement with error: %w", err)
	}

	return stmt, nil
}
