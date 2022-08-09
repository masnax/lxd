//go:build linux && cgo && !agent

package cluster

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lxc/lxd/lxd/db/query"
	"github.com/lxc/lxd/lxd/device/config"
	"github.com/lxc/lxd/lxd/instance/instancetype"
	"github.com/lxc/lxd/shared/api"
	"github.com/lxc/lxd/shared/osarch"
)

// Code generation directives.
//
//go:generate -command mapper lxd-generate db mapper -t instances.mapper.go
//go:generate mapper reset -i -b "//go:build linux && cgo && !agent"
//
//go:generate mapper stmt -e instance objects
//go:generate mapper stmt -e instance objects-by-ID
//go:generate mapper stmt -e instance objects-by-Project
//go:generate mapper stmt -e instance objects-by-Project-and-Type
//go:generate mapper stmt -e instance objects-by-Project-and-Type-and-Node
//go:generate mapper stmt -e instance objects-by-Project-and-Type-and-Node-and-Name
//go:generate mapper stmt -e instance objects-by-Project-and-Type-and-Name
//go:generate mapper stmt -e instance objects-by-Project-and-Name
//go:generate mapper stmt -e instance objects-by-Project-and-Name-and-Node
//go:generate mapper stmt -e instance objects-by-Project-and-Node
//go:generate mapper stmt -e instance objects-by-Type
//go:generate mapper stmt -e instance objects-by-Type-and-Name
//go:generate mapper stmt -e instance objects-by-Type-and-Name-and-Node
//go:generate mapper stmt -e instance objects-by-Type-and-Node
//go:generate mapper stmt -e instance objects-by-Node
//go:generate mapper stmt -e instance objects-by-Node-and-Name
//go:generate mapper stmt -e instance objects-by-Name num_filters=200
//go:generate mapper stmt -e instance id
//go:generate mapper stmt -e instance create
//go:generate mapper stmt -e instance rename
//go:generate mapper stmt -e instance delete-by-Project-and-Name
//go:generate mapper stmt -e instance update
//
//go:generate mapper method -i -e instance GetMany references=Config,Device maw=maw
//go:generate mapper method -i -e instance GetOne
//go:generate mapper method -i -e instance ID
//go:generate mapper method -i -e instance Exists
//go:generate mapper method -i -e instance Create references=Config,Device
//go:generate mapper method -i -e instance Rename
//go:generate mapper method -i -e instance DeleteOne-by-Project-and-Name
//go:generate mapper method -i -e instance Update references=Config,Device

// Instance is a value object holding db-related details about an instance.
type Instance struct {
	ID           int
	Project      string `db:"primary=yes&join=projects.name"`
	Name         string `db:"primary=yes"`
	Node         string `db:"join=nodes.name"`
	Type         instancetype.Type
	Snapshot     bool `db:"ignore"`
	Architecture int
	Ephemeral    bool
	CreationDate time.Time
	Stateful     bool
	LastUseDate  sql.NullTime
	Description  string `db:"coalesce=''"`
	ExpiryDate   sql.NullTime
}

// InstanceFilter specifies potential query parameter fields.
type InstanceFilter struct {
	ID      *int
	Project *string
	Name    *string
	Node    *string
	Type    *instancetype.Type
}

// ToAPI converts the database Instance to API type.
func (i *Instance) ToAPI(ctx context.Context, tx *sql.Tx) (*api.Instance, error) {
	profiles, err := GetInstanceProfiles(ctx, tx, i.ID)
	if err != nil {
		return nil, err
	}

	apiProfiles := make([]api.Profile, 0, len(profiles))
	profileNames := make([]string, 0, len(profiles))
	for _, p := range profiles {
		apiProfile, err := p.ToAPI(ctx, tx)
		if err != nil {
			return nil, err
		}

		apiProfiles = append(apiProfiles, *apiProfile)
		profileNames = append(profileNames, p.Name)
	}

	devices, err := GetInstanceDevices(ctx, tx, i.ID)
	if err != nil {
		return nil, err
	}

	apiDevices := DevicesToAPI(devices)
	expandedDevices := ExpandInstanceDevices(config.NewDevices(apiDevices), apiProfiles)

	config, err := GetInstanceConfig(ctx, tx, i.ID)
	if err != nil {
		return nil, err
	}

	expandedConfig := ExpandInstanceConfig(config, apiProfiles)

	archName, err := osarch.ArchitectureName(i.Architecture)
	if err != nil {
		return nil, err
	}

	return &api.Instance{
		InstancePut: api.InstancePut{
			Architecture: archName,
			Config:       config,
			Devices:      apiDevices,
			Ephemeral:    i.Ephemeral,
			Profiles:     profileNames,
			Stateful:     i.Stateful,
			Description:  i.Description,
		},
		CreatedAt:       i.CreationDate,
		ExpandedConfig:  expandedConfig,
		ExpandedDevices: expandedDevices.CloneNative(),
		Name:            i.Name,
		LastUsedAt:      i.LastUseDate.Time,
		Location:        i.Node,
		Type:            i.Type.String(),
		Project:         i.Project,
	}, nil
}

var BAW = `
SELECT instances.id, projects.name AS project, instances.name, nodes.name AS node, instances.type, instances.architecture, instances.ephemeral, instances.creation_date, instances.stateful, instances.last_use_date, coalesce(instances.description, ''), instances.expiry_date
  FROM instances JOIN projects ON instances.project_id = projects.id JOIN nodes ON instances.node_id = nodes.id
  WHERE instances.name IN ({num_filters}) ORDER BY projects.id, instances.name
`

// GetInstances returns all available instances.
// generator: instance GetMany
func GetInstances2(ctx context.Context, tx *sql.Tx, filters ...InstanceFilter) ([]Instance, error) {
	var err error

	// Result slice.
	objects := make([]Instance, 0)

	// Pick the prepared statement and arguments to use based on active criteria.
	var sqlStmt *sql.Stmt
	var args []any
	var maw string

	if len(filters) == 0 {
		sqlStmt = Stmt(tx, instanceObjects)
		args = []any{}
	}

	var ids []any
	var projects []any
	var names []any
	var nodes []any
	var types []any

	for _, filter := range filters {
		if filter.ID != nil {
			if ids == nil {
				ids = []any{}
			}

			ids = append(ids, filter.ID)
		}

		if filter.Project != nil {
			if projects == nil {
				projects = []any{}
			}

			projects = append(projects, filter.Project)
		}

		if filter.Name != nil {
			if names == nil {
				names = []any{}
			}

			names = append(names, filter.Name)
		}

		if filter.Node != nil {
			if nodes == nil {
				nodes = []any{}
			}

			nodes = append(nodes, filter.Node)
		}

		if filter.Type != nil {
			if types == nil {
				types = []any{}
			}

			types = append(types, filter.Type)
		}

	}

	if len(filters) > 0 {
		if projects != nil && types != nil && nodes != nil && names != nil && ids == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndTypeAndNodeAndName]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndTypeAndNodeAndName)
			}

			args = []any{projects, types, nodes, names}
		} else if projects != nil && types != nil && nodes != nil && ids == nil && names == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndTypeAndNode]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndTypeAndNode)
			}

			args = []any{projects, types, nodes}
		} else if projects != nil && types != nil && names != nil && ids == nil && nodes == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndTypeAndName]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndTypeAndName)
			}

			args = []any{projects, types, names}
		} else if types != nil && names != nil && nodes != nil && ids == nil && projects == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByTypeAndNameAndNode]
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByTypeAndNameAndNode)
			}

			args = []any{types, names, nodes}
		} else if projects != nil && names != nil && nodes != nil && ids == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndNameAndNode]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndNameAndNode)
			}

			args = []any{projects, names, nodes}
		} else if projects != nil && types != nil && ids == nil && names == nil && nodes == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndType]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndType)
			}

			args = []any{projects, types}
		} else if types != nil && nodes != nil && ids == nil && projects == nil && names == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByTypeAndNode]
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByTypeAndNode)
			}

			args = []any{types, nodes}
		} else if types != nil && names != nil && ids == nil && projects == nil && nodes == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByTypeAndName]
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByTypeAndName)
			}

			args = []any{types, names}
		} else if projects != nil && nodes != nil && ids == nil && names == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndNode]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndNode)
			}

			args = []any{projects, nodes}
		} else if projects != nil && names != nil && ids == nil && nodes == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProjectAndName]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProjectAndName)
			}

			args = []any{projects, names}
		} else if nodes != nil && names != nil && ids == nil && projects == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByNodeAndName]
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByNodeAndName)
			}

			args = []any{nodes, names}
		} else if types != nil && ids == nil && projects == nil && names == nil && nodes == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByType]
				queryStr = strings.Replace(query.query, "type = ?", fmt.Sprintf("type IN (?%s)", strings.Repeat(", ?", len(types)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByType)
			}

			args = []any{types}
		} else if projects != nil && ids == nil && names == nil && nodes == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByProject]
				queryStr = strings.Replace(query.query, "project = ?", fmt.Sprintf("project IN (?%s)", strings.Repeat(", ?", len(projects)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByProject)
			}

			args = []any{projects}
		} else if nodes != nil && ids == nil && projects == nil && names == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByNode]
				queryStr = strings.Replace(query.query, "node = ?", fmt.Sprintf("node IN (?%s)", strings.Repeat(", ?", len(nodes)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByNode)
			}

			args = []any{nodes}
		} else if names != nil && ids == nil && projects == nil && nodes == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByName]
				queryStr = strings.Replace(query.query, "name = ?", fmt.Sprintf("name IN (?%s)", strings.Repeat(", ?", len(names)-1)), -1)

				maw = queryStr
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByName)
			}

			args = []any{names}
		} else if ids != nil && projects == nil && names == nil && nodes == nil && types == nil {
			if len(filters) > 1 {
				var queryStr string
				query := stmts[instanceObjectsByID]
				queryStr = strings.Replace(query.query, "id = ?", fmt.Sprintf("id IN (?%s)", strings.Repeat(", ?", len(ids)-1)), -1)

				sqlStmt, err = prepare(tx, queryStr)
				if err != nil {
					return nil, fmt.Errorf("Failed to prepare stmt: %w", err)
				}
			} else {
				sqlStmt = Stmt(tx, instanceObjectsByID)
			}

			args = []any{ids}
		} else if ids == nil && projects == nil && names == nil && nodes == nil && types == nil {
			sqlStmt = Stmt(tx, instanceObjects)
			args = []any{}
		} else {
			return nil, fmt.Errorf("No statement exists for the given Filter")
		}
	}

	// Dest function for scanning a row.
	dest := func(i int) []any {
		objects = append(objects, Instance{})
		return []any{
			&objects[i].ID,
			&objects[i].Project,
			&objects[i].Name,
			&objects[i].Node,
			&objects[i].Type,
			&objects[i].Architecture,
			&objects[i].Ephemeral,
			&objects[i].CreationDate,
			&objects[i].Stateful,
			&objects[i].LastUseDate,
			&objects[i].Description,
			&objects[i].ExpiryDate,
		}
	}

	// Select.
	allArgs := []any{}
	for _, arg := range args {
		allArgs = append(allArgs, arg.([]any)...)
	}

	if maw != "" {
		err = query.QueryObjects(tx, maw, dest, allArgs...)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch from \"instances\" table: %w", err)
		}

	} else {
		err = query.SelectObjects(sqlStmt, dest, allArgs...)
		if err != nil {
			return nil, fmt.Errorf("Failed to fetch from \"instances\" table: %w", err)
		}

	}
	return objects, nil
}
