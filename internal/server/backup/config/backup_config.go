package config

import (
	"errors"
	"fmt"

	internalInstance "github.com/lxc/incus/v7/internal/instance"
	"github.com/lxc/incus/v7/shared/api"
)

// Volume is a volume with its snapshots as listed in the LXD version 2 metadata.
type Volume struct {
	api.StorageVolume `yaml:",inline"`

	Snapshots []*api.StorageVolumeSnapshot `yaml:"snapshots,omitempty"`
}

type configAlias Config

// rawConfig is Config plus the LXD version 2 keys, which are mapped to the Config fields on load.
type rawConfig struct {
	configAlias `yaml:",inline"`

	Version  int                `yaml:"version,omitempty"`
	Instance *api.Instance      `yaml:"instance,omitempty"`
	Pools    []*api.StoragePool `yaml:"pools,omitempty"`
	Volumes  []*Volume          `yaml:"volumes,omitempty"`
}

// Config represents the config of a backup that can be stored in a backup.yaml file (or embedded in index.yaml).
type Config struct {
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

// UnmarshalYAML loads version 1 metadata as is and maps LXD version 2 metadata onto the same fields.
func (c *Config) UnmarshalYAML(unmarshal func(any) error) error {
	raw := rawConfig{}
	err := unmarshal(&raw)
	if err != nil {
		return err
	}

	*c = Config(raw.configAlias)

	switch raw.Version {
	case 0, 1:
		return nil
	case 2:
		return c.fromV2(&raw)
	}

	return fmt.Errorf("Unsupported backup metadata version %d", raw.Version)
}

func (c *Config) fromV2(raw *rawConfig) error {
	for _, pool := range raw.Pools {
		if pool == nil {
			return errors.New("Backup metadata contains a nil pool")
		}
	}

	for _, vol := range raw.Volumes {
		if vol == nil {
			return errors.New("Backup metadata contains a nil volume")
		}
	}

	var vol *Volume
	var poolName string

	if raw.Instance != nil {
		c.Container = raw.Instance

		// The root volume has the instance's name and isn't a custom volume.
		for _, v := range raw.Volumes {
			if v.Name == raw.Instance.Name && v.Type != "custom" {
				vol = v
				break
			}
		}

		_, rootDev, err := internalInstance.GetRootDiskDevice(raw.Instance.ExpandedDevices)
		if err != nil {
			_, rootDev, _ = internalInstance.GetRootDiskDevice(raw.Instance.Devices)
		}

		poolName = rootDev["pool"]
	} else if len(raw.Volumes) > 1 {
		return errors.New("More than one primary volume is defined in backup metadata")
	} else if len(raw.Volumes) == 1 {
		vol = raw.Volumes[0]
	}

	for _, pool := range raw.Pools {
		if pool.Name == poolName {
			c.Pool = pool
			break
		}
	}

	if c.Pool == nil && len(raw.Pools) == 1 {
		c.Pool = raw.Pools[0]
	}

	if vol != nil {
		c.Volume = &vol.StorageVolume
		c.VolumeSnapshots = vol.Snapshots
	}

	return nil
}
