//go:build linux && cgo && !agent

package cluster

// The code below was generated by lxd-generate - DO NOT EDIT!

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/shared/api"
)

var _ = api.ServerEnvironment{}

var operationObjects = RegisterStmt(`
SELECT operations.id, operations.uuid, nodes.address AS node_address, operations.project_id, operations.node_id, operations.type
  FROM operations JOIN nodes ON operations.node_id = nodes.id
  ORDER BY operations.id, operations.uuid
`)

var operationObjectsByNodeID = RegisterStmt(`
SELECT operations.id, operations.uuid, nodes.address AS node_address, operations.project_id, operations.node_id, operations.type
  FROM operations JOIN nodes ON operations.node_id = nodes.id
  WHERE operations.node_id = ? ORDER BY operations.id, operations.uuid
`)

var operationObjectsByID = RegisterStmt(`
SELECT operations.id, operations.uuid, nodes.address AS node_address, operations.project_id, operations.node_id, operations.type
  FROM operations JOIN nodes ON operations.node_id = nodes.id
  WHERE operations.id = ? ORDER BY operations.id, operations.uuid
`)

var operationObjectsByUUID = RegisterStmt(`
SELECT operations.id, operations.uuid, nodes.address AS node_address, operations.project_id, operations.node_id, operations.type
  FROM operations JOIN nodes ON operations.node_id = nodes.id
  WHERE operations.uuid = ? ORDER BY operations.id, operations.uuid
`)

var operationCreateOrReplace = RegisterStmt(`
INSERT OR REPLACE INTO operations (uuid, project_id, node_id, type)
 VALUES (?, ?, ?, ?)
`)

var operationDeleteByUUID = RegisterStmt(`
DELETE FROM operations WHERE uuid = ?
`)

var operationDeleteByNodeID = RegisterStmt(`
DELETE FROM operations WHERE node_id = ?
`)

// GetOperations returns all available operations.
// generator: operation GetMany
func GetOperations(ctx context.Context, tx *sql.Tx, filter OperationFilter) ([]Operation, error) {
	var err error

	// Result slice.
	objects := make([]Operation, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	var sqlStmt *sql.Stmt
	var args []any

	if filter.UUID != nil && filter.ID == nil && filter.NodeID == nil {
		sqlStmt, err = Stmt(tx, operationObjectsByUUID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get \"operationObjectsByUUID\" prepared statement: %w", err)
		}

		args = []any{
			filter.UUID,
		}
	} else if filter.NodeID != nil && filter.ID == nil && filter.UUID == nil {
		sqlStmt, err = Stmt(tx, operationObjectsByNodeID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get \"operationObjectsByNodeID\" prepared statement: %w", err)
		}

		args = []any{
			filter.NodeID,
		}
	} else if filter.ID != nil && filter.NodeID == nil && filter.UUID == nil {
		sqlStmt, err = Stmt(tx, operationObjectsByID)
		if err != nil {
			return nil, fmt.Errorf("Failed to get \"operationObjectsByID\" prepared statement: %w", err)
		}

		args = []any{
			filter.ID,
		}
	} else if filter.ID == nil && filter.NodeID == nil && filter.UUID == nil {
		sqlStmt, err = Stmt(tx, operationObjects)
		if err != nil {
			return nil, fmt.Errorf("Failed to get \"operationObjects\" prepared statement: %w", err)
		}

		args = []any{}
	} else {
		return nil, fmt.Errorf("No statement exists for the given Filter")
	}

	// Dest function for scanning a row.
	dest := func(scan func(dest ...any) error) error {
		o := Operation{}
		err := scan(&o.ID, &o.UUID, &o.NodeAddress, &o.ProjectID, &o.NodeID, &o.Type)
		if err != nil {
			return err
		}

		objects = append(objects, o)

		return nil
	}

	// Select.
	err = query.SelectObjects(sqlStmt, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"operations\" table: %w", err)
	}

	return objects, nil
}

// CreateOrReplaceOperation adds a new operation to the database.
// generator: operation CreateOrReplace
func CreateOrReplaceOperation(ctx context.Context, tx *sql.Tx, object Operation) (int64, error) {
	args := make([]any, 4)

	// Populate the statement arguments.
	args[0] = object.UUID
	args[1] = object.ProjectID
	args[2] = object.NodeID
	args[3] = object.Type

	// Prepared statement to use.
	stmt, err := Stmt(tx, operationCreateOrReplace)
	if err != nil {
		return -1, fmt.Errorf("Failed to get \"operationCreateOrReplace\" prepared statement: %w", err)
	}

	// Execute the statement.
	result, err := stmt.Exec(args...)
	if err != nil {
		return -1, fmt.Errorf("Failed to create \"operations\" entry: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("Failed to fetch \"operations\" entry ID: %w", err)
	}

	return id, nil
}

// DeleteOperation deletes the operation matching the given key parameters.
// generator: operation DeleteOne-by-UUID
func DeleteOperation(ctx context.Context, tx *sql.Tx, uuid string) error {
	stmt, err := Stmt(tx, operationDeleteByUUID)
	if err != nil {
		return fmt.Errorf("Failed to get \"operationDeleteByUUID\" prepared statement: %w", err)
	}

	result, err := stmt.Exec(uuid)
	if err != nil {
		return fmt.Errorf("Delete \"operations\": %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "Operation not found")
	} else if n > 1 {
		return fmt.Errorf("Query deleted %d Operation rows instead of 1", n)
	}

	return nil
}

// DeleteOperations deletes the operation matching the given key parameters.
// generator: operation DeleteMany-by-NodeID
func DeleteOperations(ctx context.Context, tx *sql.Tx, nodeID int64) error {
	stmt, err := Stmt(tx, operationDeleteByNodeID)
	if err != nil {
		return fmt.Errorf("Failed to get \"operationDeleteByNodeID\" prepared statement: %w", err)
	}

	result, err := stmt.Exec(nodeID)
	if err != nil {
		return fmt.Errorf("Delete \"operations\": %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	return nil
}
