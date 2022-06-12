//go:build linux && cgo && !agent

package cluster

import (
	"context"
	"database/sql"

	"github.com/lxc/lxd/lxd/device/config"
	"github.com/lxc/lxd/shared/api"
)

// Code generation directives.
//
//go:generate -command mapper lxd-generate db mapper -t profiles.mapper.go
//go:generate mapper reset -i -b "//go:build linux && cgo && !agent"
//
//go:generate mapper stmt -e profile objects version=2
//go:generate mapper stmt -e profile objects-by-ID version=2
//go:generate mapper stmt -e profile objects-by-Name version=2
//go:generate mapper stmt -e profile objects-by-Project version=2
//go:generate mapper stmt -e profile objects-by-Project-and-Name version=2
//go:generate mapper stmt -e profile id version=2
//go:generate mapper stmt -e profile create version=2
//go:generate mapper stmt -e profile rename version=2
//go:generate mapper stmt -e profile update version=2
//go:generate mapper stmt -e profile delete-by-Project-and-Name version=2
//
//go:generate mapper method -i -e profile ID version=2
//go:generate mapper method -i -e profile Exists version=2
//go:generate mapper method -i -e profile GetMany references=Config,Device version=2
//go:generate mapper method -i -e profile GetOne version=2
//go:generate mapper method -i -e profile Create references=Config,Device version=2
//go:generate mapper method -i -e profile Rename version=2
//go:generate mapper method -i -e profile Update references=Config,Device version=2
//go:generate mapper method -i -e profile DeleteOne-by-Project-and-Name version=2

// Profile is a value object holding db-related details about a profile.
type Profile struct {
	ID          int
	ProjectID   int    `db:"omit=create,update"`
	Project     string `db:"primary=yes&join=projects.name"`
	Name        string `db:"primary=yes"`
	Description string `db:"coalesce=''"`
}

// ProfileFilter specifies potential query parameter fields.
type ProfileFilter struct {
	ID      *int
	Project *string
	Name    *string
}

// ToAPI returns a cluster Profile as an API struct.
func (p *Profile) ToAPI(ctx context.Context, tx *sql.Tx) (*api.Profile, error) {
	config, err := GetProfileConfig(ctx, tx, p.ID)
	if err != nil {
		return nil, err
	}

	devices, err := GetProfileDevices(ctx, tx, p.ID)
	if err != nil {
		return nil, err
	}

	profile := &api.Profile{
		Name: p.Name,
		ProfilePut: api.ProfilePut{
			Description: p.Description,
			Config:      config,
			Devices:     DevicesToAPI(devices),
		},
	}

	return profile, nil
}

// GetProfileIfEnabled returns the profile from the given project, or the
// default project if "features.profiles" is not set.
func GetProfileIfEnabled(ctx context.Context, tx *sql.Tx, projectName string, name string) (*Profile, error) {
	enabled, err := ProjectHasProfiles(ctx, tx, projectName)
	if err != nil {
		return nil, err
	}

	if !enabled {
		projectName = "default"
	}

	return GetProfile(ctx, tx, projectName, name)
}

// ExpandInstanceConfig expands the given instance config with the config
// values of the given profiles.
func ExpandInstanceConfig(config map[string]string, profiles []api.Profile) map[string]string {
	expandedConfig := map[string]string{}

	// Apply all the profiles
	profileConfigs := make([]map[string]string, len(profiles))
	for i, profile := range profiles {
		profileConfigs[i] = profile.Config
	}

	for i := range profileConfigs {
		for k, v := range profileConfigs[i] {
			expandedConfig[k] = v
		}
	}

	// Stick the given config on top
	for k, v := range config {
		expandedConfig[k] = v
	}

	return expandedConfig
}

// ExpandInstanceDevices expands the given instance devices with the devices
// defined in the given profiles.
func ExpandInstanceDevices(devices config.Devices, profiles []api.Profile) config.Devices {
	expandedDevices := config.Devices{}

	// Apply all the profiles
	profileDevices := make([]config.Devices, len(profiles))
	for i, profile := range profiles {
		profileDevices[i] = config.NewDevices(profile.Devices)
	}
	for i := range profileDevices {
		for k, v := range profileDevices[i] {
			expandedDevices[k] = v
		}
	}

	// Stick the given devices on top
	for k, v := range devices {
		expandedDevices[k] = v
	}

	return expandedDevices
}
