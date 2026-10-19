package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"
)

const lxdV2Container = `
version: 2
instance:
  name: c1
  type: container
  expanded_devices:
    root: {type: disk, path: /, pool: default}
pools:
- name: default
  driver: dir
volumes:
- name: c1
  type: container
  pool: default
  snapshots:
  - name: snap0
- name: data
  type: custom
  pool: default
`

const lxdV2VM = `
version: 2
instance:
  name: vm1
  type: virtual-machine
pools:
- name: default
  driver: zfs
volumes:
- name: vm1
  type: virtual-machine
  content_type: block
  pool: default
`

const lxdV2Custom = `
version: 2
pools:
- name: default
  driver: dir
volumes:
- name: data
  type: custom
  pool: default
  snapshots:
  - name: s0
  - name: s1
`

const v1Container = `
container:
  name: c1
  type: container
pool:
  name: default
volume:
  name: c1
volume_snapshots:
- name: snap0
`

func normalize(t *testing.T, data string) (*Config, error) {
	t.Helper()

	c := &Config{}
	err := yaml.Load([]byte(data), c)
	require.NoError(t, err)

	return c, c.Normalize()
}

func TestNormalizeV2Container(t *testing.T) {
	c, err := normalize(t, lxdV2Container)
	require.NoError(t, err)
	require.NotNil(t, c.Container)
	require.Equal(t, "c1", c.Container.Name)
	require.Equal(t, "default", c.Pool.Name)
	require.Equal(t, "c1", c.Volume.Name)
	require.Len(t, c.VolumeSnapshots, 1)
	require.Equal(t, "snap0", c.VolumeSnapshots[0].Name)
	require.Nil(t, c.Instance)
	require.Nil(t, c.Volumes)
	require.Zero(t, c.Version)
}

func TestNormalizeV2VM(t *testing.T) {
	c, err := normalize(t, lxdV2VM)
	require.NoError(t, err)
	require.Equal(t, "virtual-machine", c.Container.Type)
	require.Equal(t, "zfs", c.Pool.Driver)
	require.Equal(t, "block", c.Volume.ContentType)
}

func TestNormalizeV2Custom(t *testing.T) {
	c, err := normalize(t, lxdV2Custom)
	require.NoError(t, err)
	require.Nil(t, c.Container)
	require.Equal(t, "data", c.Volume.Name)
	require.Len(t, c.VolumeSnapshots, 2)
}

func TestNormalizeV1Untouched(t *testing.T) {
	c, err := normalize(t, v1Container)
	require.NoError(t, err)
	require.Equal(t, "c1", c.Container.Name)
	require.Len(t, c.VolumeSnapshots, 1)
}

func TestNormalizeUnsupportedVersion(t *testing.T) {
	_, err := normalize(t, "version: 3\ninstance:\n  name: c1\n")
	require.ErrorContains(t, err, "Unsupported backup metadata version 3")
}

func TestNormalizeNilVolume(t *testing.T) {
	c := &Config{Version: MetadataVersion2, Volumes: []*Volume{nil}}
	require.Error(t, c.Normalize())
}

func TestNormalizeNilReceiver(t *testing.T) {
	var c *Config
	require.NoError(t, c.Normalize())
}
