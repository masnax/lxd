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

var warningObjects = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  ORDER BY warnings.uuid
`)

var warningObjectsByUUID = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE warnings.uuid = ? ORDER BY warnings.uuid
`)

var warningObjectsByProject = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE coalesce(project, '') = ? ORDER BY warnings.uuid
`)

var warningObjectsByStatus = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE warnings.status = ? ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCode = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE coalesce(node, '') = ? AND warnings.type_code = ? ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCodeAndProject = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE coalesce(node, '') = ? AND warnings.type_code = ? AND coalesce(project, '') = ? ORDER BY warnings.uuid
`)

var warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID = RegisterStmt(`
SELECT warnings.id, coalesce(nodes.name, '') AS node, coalesce(projects.name, '') AS project, coalesce(warnings.entity_type_code, -1), coalesce(warnings.entity_id, -1), warnings.uuid, warnings.type_code, warnings.status, warnings.first_seen_date, warnings.last_seen_date, warnings.updated_date, warnings.last_message, warnings.count
  FROM warnings LEFT JOIN nodes ON warnings.node_id = nodes.id LEFT JOIN projects ON warnings.project_id = projects.id
  WHERE coalesce(node, '') = ? AND warnings.type_code = ? AND coalesce(project, '') = ? AND coalesce(warnings.entity_type_code, -1) = ? AND coalesce(warnings.entity_id, -1) = ? ORDER BY warnings.uuid
`)

var warningDeleteByUUID = RegisterStmt(`
DELETE FROM warnings WHERE uuid = ?
`)

var warningDeleteByEntityTypeCodeAndEntityID = RegisterStmt(`
DELETE FROM warnings WHERE entity_type_code = ? AND entity_id = ?
`)

var warningID = RegisterStmt(`
SELECT warnings.id FROM warnings
  WHERE warnings.uuid = ?
`)

// GetWarnings returns all available warnings.
// generator: warning GetMany
func GetWarnings(ctx context.Context, tx *sql.Tx, filter WarningFilter) ([]Warning, error) {
	var err error

	// Result slice.
	objects := make([]Warning, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	var sqlStmt *sql.Stmt
	var args []any

	if filter.Node != nil && filter.TypeCode != nil && filter.Project != nil && filter.EntityTypeCode != nil && filter.EntityID != nil && filter.ID == nil && filter.UUID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjectsByNodeAndTypeCodeAndProjectAndEntityTypeCodeAndEntityID)
		args = []any{
			filter.Node,
			filter.TypeCode,
			filter.Project,
			filter.EntityTypeCode,
			filter.EntityID,
		}
	} else if filter.Node != nil && filter.TypeCode != nil && filter.Project != nil && filter.ID == nil && filter.UUID == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjectsByNodeAndTypeCodeAndProject)
		args = []any{
			filter.Node,
			filter.TypeCode,
			filter.Project,
		}
	} else if filter.Node != nil && filter.TypeCode != nil && filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjectsByNodeAndTypeCode)
		args = []any{
			filter.Node,
			filter.TypeCode,
		}
	} else if filter.UUID != nil && filter.ID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjectsByUUID)
		args = []any{
			filter.UUID,
		}
	} else if filter.Status != nil && filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil {
		sqlStmt = Stmt(tx, warningObjectsByStatus)
		args = []any{
			filter.Status,
		}
	} else if filter.Project != nil && filter.ID == nil && filter.UUID == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjectsByProject)
		args = []any{
			filter.Project,
		}
	} else if filter.ID == nil && filter.UUID == nil && filter.Project == nil && filter.Node == nil && filter.TypeCode == nil && filter.EntityTypeCode == nil && filter.EntityID == nil && filter.Status == nil {
		sqlStmt = Stmt(tx, warningObjects)
		args = []any{}
	} else {
		return nil, fmt.Errorf("No statement exists for the given Filter")
	}

	// Dest function for scanning a row.
	dest := func(scan func(dest ...any) error) error {
		w := Warning{}
		err := scan(&w.ID, &w.Node, &w.Project, &w.EntityTypeCode, &w.EntityID, &w.UUID, &w.TypeCode, &w.Status, &w.FirstSeenDate, &w.LastSeenDate, &w.UpdatedDate, &w.LastMessage, &w.Count)
		if err != nil {
			return err
		}

		objects = append(objects, w)

		return nil
	}

	// Select.
	err = query.SelectObjects(sqlStmt, dest, args...)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	return objects, nil
}

// GetWarning returns the warning with the given key.
// generator: warning GetOne-by-UUID
func GetWarning(ctx context.Context, tx *sql.Tx, uuid string) (*Warning, error) {
	filter := WarningFilter{}
	filter.UUID = &uuid

	objects, err := GetWarnings(ctx, tx, filter)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch from \"warnings\" table: %w", err)
	}

	switch len(objects) {
	case 0:
		return nil, api.StatusErrorf(http.StatusNotFound, "Warning not found")
	case 1:
		return &objects[0], nil
	default:
		return nil, fmt.Errorf("More than one \"warnings\" entry matches")
	}
}

// DeleteWarning deletes the warning matching the given key parameters.
// generator: warning DeleteOne-by-UUID
func DeleteWarning(ctx context.Context, tx *sql.Tx, uuid string) error {
	stmt := Stmt(tx, warningDeleteByUUID)
	result, err := stmt.Exec(uuid)
	if err != nil {
		return fmt.Errorf("Delete \"warnings\": %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	if n == 0 {
		return api.StatusErrorf(http.StatusNotFound, "Warning not found")
	} else if n > 1 {
		return fmt.Errorf("Query deleted %d Warning rows instead of 1", n)
	}

	return nil
}

// DeleteWarnings deletes the warning matching the given key parameters.
// generator: warning DeleteMany-by-EntityTypeCode-and-EntityID
func DeleteWarnings(ctx context.Context, tx *sql.Tx, entityTypeCode int, entityID int) error {
	stmt := Stmt(tx, warningDeleteByEntityTypeCodeAndEntityID)
	result, err := stmt.Exec(entityTypeCode, entityID)
	if err != nil {
		return fmt.Errorf("Delete \"warnings\": %w", err)
	}

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	return nil
}

// GetWarningID return the ID of the warning with the given key.
// generator: warning ID
func GetWarningID(ctx context.Context, tx *sql.Tx, uuid string) (int64, error) {
	stmt := Stmt(tx, warningID)
	rows, err := stmt.Query(uuid)
	if err != nil {
		return -1, fmt.Errorf("Failed to get \"warnings\" ID: %w", err)
	}

	defer func() { _ = rows.Close() }()

	// Ensure we read one and only one row.
	if !rows.Next() {
		return -1, api.StatusErrorf(http.StatusNotFound, "Warning not found")
	}

	var id int64
	err = rows.Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("Failed to scan ID: %w", err)
	}

	if rows.Next() {
		return -1, fmt.Errorf("More than one row returned")
	}

	err = rows.Err()
	if err != nil {
		return -1, fmt.Errorf("Result set failure: %w", err)
	}

	return id, nil
}

// WarningExists checks if a warning with the given key exists.
// generator: warning Exists
func WarningExists(ctx context.Context, tx *sql.Tx, uuid string) (bool, error) {
	_, err := GetWarningID(ctx, tx, uuid)
	if err != nil {
		if api.StatusErrorCheck(err, http.StatusNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
