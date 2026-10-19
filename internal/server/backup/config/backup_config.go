package config

import (
	"errors"
	"fmt"

	"github.com/lxc/incus/v7/shared/api"
)

// MetadataVersion1 is the original backup metadata format, used by Incus and LXD before 6.x.
const MetadataVersion1 = 1

// MetadataVersion2 is the LXD 6.x format, converted to the internal format on load.
const MetadataVersion2 = 2

// Volume represents a volume with its snapshots as found in the LXD version 2 metadata.
type Volume struct {
	api.StorageVolume `yaml:",inline"`

	Snapshots []*api.StorageVolumeSnapshot `yaml:"snapshots,omitempty"`
}

// Config represents the config of a backup that can be stored in a backup.yaml file (or embedded in index.yaml).
type Config struct {
	// Fields only set when reading LXD version 2 metadata, cleared by Normalize.
	Version  int                `yaml:"version,omitempty"`
	Instance *api.Instance      `yaml:"instance,omitempty"`
	Pools    []*api.StoragePool `yaml:"pools,omitempty"`
	Volumes  []*Volume          `yaml:"volumes,omitempty"`

	Container        *api.Instance                `yaml:"container,omitempty"` // Used by VM backups too.
	Snapshots        []*api.InstanceSnapshot      `yaml:"snapshots,omitempty"`
	Pool             *api.StoragePool             `yaml:"pool,omitempty"`
	Profiles         []*api.Profile               `yaml:"profiles,omitempty"`
	Volume           *api.StorageVolume           `yaml:"volume,omitempty"`
	VolumeSnapshots  []*api.StorageVolumeSnapshot `yaml:"volume_snapshots,omitempty"`
	DependentVolumes []*Config                    `yaml:"dependent_volumes,omitempty"`
	Bucket           *api.StorageBucket           `yaml:"bucket,omitempty"`
	BucketKeys       []*api.StorageBucketKey      `yaml:"bucket_keys,omitempty"`
}

// Normalize converts LXD version 2 metadata to the internal (version 1) layout.
// Mapping follows ConvertFormat's downgrade path in canonical/lxd lxd/backup/backup_config_utils.go.
func (c *Config) Normalize() error {
	if c == nil {
		return nil
	}

	if c.Version > MetadataVersion2 {
		return fmt.Errorf("Unsupported backup metadata version %d", c.Version)
	}

	if c.Instance != nil {
		c.Container = c.Instance

		if len(c.Pools) > 0 {
			c.Pool = c.Pools[0]
		}
	}

	if len(c.Volumes) > 0 {
		vol := c.Volumes[0]
		for _, v := range c.Volumes {
			if v == nil {
				return errors.New("Backup metadata contains a nil volume")
			}

			// The root volume carries the instance's name and isn't a custom volume.
			if c.Instance != nil && v.Name == c.Instance.Name && v.Type != "custom" {
				vol = v
				break
			}
		}

		if vol == nil {
			return errors.New("Backup metadata contains a nil volume")
		}

		c.Volume = &vol.StorageVolume
		c.VolumeSnapshots = vol.Snapshots
	}

	c.Version = 0
	c.Instance = nil
	c.Pools = nil
	c.Volumes = nil

	return nil
}
