package cluster

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/logger"
)

//go:generate -command mapper lxd-generate db mapper -t cluster_groups.mapper.go
//go:generate mapper reset -i -b "//go:build linux && cgo && !agent"
//
//go:generate mapper stmt -e cluster_group objects table=cluster_groups
//go:generate mapper stmt -e cluster_group objects-by-Name table=cluster_groups
//go:generate mapper stmt -e cluster_group id table=cluster_groups
//go:generate mapper stmt -e cluster_group create table=cluster_groups
//go:generate mapper stmt -e cluster_group rename table=cluster_groups
//go:generate mapper stmt -e cluster_group delete-by-Name table=cluster_groups
//go:generate mapper stmt -e cluster_group update table=cluster_groups
//
//go:generate mapper method -i -e cluster_group GetMany
//go:generate mapper method -i -e cluster_group GetOne
//go:generate mapper method -i -e cluster_group ID
//go:generate mapper method -i -e cluster_group Exists
//go:generate mapper method -i -e cluster_group Rename
//go:generate mapper method -i -e cluster_group Create
//go:generate mapper method -i -e cluster_group DeleteOne-by-Name

// ClusterGroup is a value object holding db-related details about a cluster group.
type ClusterGroup struct {
	ID          int
	Name        string
	Description string
	Nodes       []string `db:"ignore"`
}

// ClusterGroupFilter specifies potential query parameter fields.
type ClusterGroupFilter struct {
	ID   *int
	Name *string
}

// ToAPI returns a LXD API entry.
func (c *ClusterGroup) ToAPI() (*api.ClusterGroup, error) {
	result := api.ClusterGroup{
		ClusterGroupPut: api.ClusterGroupPut{
			Description: c.Description,
			Members:     c.Nodes,
		},
		ClusterGroupPost: api.ClusterGroupPost{
			Name: c.Name,
		},
	}

	return &result, nil
}

var clusterGroupDeleteNodesRef = RegisterStmt(`
DELETE FROM nodes_cluster_groups WHERE group_id = ?
`)

// UpdateClusterGroup updates the ClusterGroup matching the given key parameters.
// generator: ClusterGroup Update
func UpdateClusterGroup(ctx context.Context, tx *sql.Tx, name string, object ClusterGroup) error {
	id, err := GetClusterGroupID(ctx, tx, name)
	if err != nil {
		return fmt.Errorf("Failed to get cluster group: %w", err)
	}

	stmt, err := Stmt(tx, clusterGroupUpdate)
	if err != nil {
		return fmt.Errorf("Failed to get \"clusterGroupUpdate\" prepared statement: %w", err)
	}

	result, err := stmt.Exec(object.Name, object.Description, id)
	if err != nil {
		return fmt.Errorf("Update \"clusters_groups\" entry failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Fetch affected rows: %w", err)
	}

	if n != 1 {
		return fmt.Errorf("Query updated %d rows instead of 1", n)
	}

	// Delete current nodes.
	stmt, err = Stmt(tx, clusterGroupDeleteNodesRef)
	if err != nil {
		return fmt.Errorf("Failed to get \"clusterGroupDeleteNodesRef\" prepared statement: %w", err)
	}

	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("Failed to delete current nodes: %w", err)
	}

	// Insert nodes reference.
	err = addNodesToClusterGroup(tx, int(id), object.Nodes)
	if err != nil {
		return fmt.Errorf("Failed to insert nodes for cluster group: %w", err)
	}

	return nil
}

// addNodesToClusterGroup adds the given nodes the the cluster group with the given ID.
func addNodesToClusterGroup(tx *sql.Tx, id int, nodes []string) error {
	str := `
INSERT INTO nodes_cluster_groups (group_id, node_id)
  VALUES (
    ?,
    (SELECT nodes.id
     FROM nodes
     WHERE nodes.name = ?)
  )`
	stmt, err := tx.Prepare(str)
	if err != nil {
		return err
	}

	defer func() { _ = stmt.Close() }()

	for _, node := range nodes {
		_, err = stmt.Exec(id, node)
		if err != nil {
			logger.Debugf("Error adding node %q to cluster group: %s", node, err)
			return err
		}
	}

	return nil
}
