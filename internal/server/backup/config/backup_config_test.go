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
- name: other
  driver: zfs
- name: default
  driver: dir
volumes:
- name: data
  type: custom
- name: c1
  type: container
  snapshots:
  - name: snap0
`

const lxdV2VM = `
version: 2
instance:
  name: vm1
  type: virtual-machine
  devices:
    root: {type: disk, path: /, pool: default}
pools:
- name: default
  driver: zfs
volumes:
- name: vm1
  type: virtual-machine
  content_type: block
`

const lxdV2Custom = `
version: 2
pools:
- name: default
  driver: dir
volumes:
- name: data
  type: custom
  snapshots:
  - name: s0
  - name: s1
`

const v1Container = `
container:
  name: c1
pool:
  name: default
volume:
  name: c1
volume_snapshots:
- name: snap0
dependent_volumes:
- volume:
    name: dep
`

func load(t *testing.T, data string) (*Config, error) {
	t.Helper()

	c := &Config{}
	err := yaml.Load([]byte(data), c)

	return c, err
}

func TestLXDv2Container(t *testing.T) {
	c, err := load(t, lxdV2Container)
	require.NoError(t, err)
	require.Equal(t, "c1", c.Container.Name)
	require.Equal(t, "default", c.Pool.Name)
	require.Equal(t, "c1", c.Volume.Name)
	require.Len(t, c.VolumeSnapshots, 1)
	require.Equal(t, "snap0", c.VolumeSnapshots[0].Name)
}

func TestLXDv2VM(t *testing.T) {
	c, err := load(t, lxdV2VM)
	require.NoError(t, err)
	require.Equal(t, "virtual-machine", c.Container.Type)
	require.Equal(t, "zfs", c.Pool.Driver)
	require.Equal(t, "block", c.Volume.ContentType)
}

func TestLXDv2Custom(t *testing.T) {
	c, err := load(t, lxdV2Custom)
	require.NoError(t, err)
	require.Nil(t, c.Container)
	require.Equal(t, "dir", c.Pool.Driver)
	require.Equal(t, "data", c.Volume.Name)
	require.Len(t, c.VolumeSnapshots, 2)
}

func TestLXDv2NoRootVolume(t *testing.T) {
	c, err := load(t, "version: 2\ninstance:\n  name: c1\nvolumes:\n- name: data\n  type: custom\n")
	require.NoError(t, err)
	require.Nil(t, c.Volume)
}

func TestV1(t *testing.T) {
	c, err := load(t, v1Container)
	require.NoError(t, err)
	require.Equal(t, "c1", c.Container.Name)
	require.Len(t, c.VolumeSnapshots, 1)
	require.Equal(t, "dep", c.DependentVolumes[0].Volume.Name)
}

func TestV1IgnoresV2Keys(t *testing.T) {
	c, err := load(t, "version: 1\ncontainer:\n  name: c1\ninstance:\n  name: other\n")
	require.NoError(t, err)
	require.Equal(t, "c1", c.Container.Name)
}

func TestInvalidVersion(t *testing.T) {
	for _, v := range []string{"3", "-1"} {
		_, err := load(t, "version: "+v+"\ninstance:\n  name: c1\n")
		require.ErrorContains(t, err, "Unsupported backup metadata version "+v)
	}
}

func TestLXDv2NilVolume(t *testing.T) {
	_, err := load(t, "version: 2\nvolumes:\n- null\n")
	require.Error(t, err)
}

func TestLXDv2MultipleCustomVolumes(t *testing.T) {
	_, err := load(t, "version: 2\nvolumes:\n- name: a\n- name: b\n")
	require.Error(t, err)
}

func TestMarshalUnchanged(t *testing.T) {
	c, err := load(t, lxdV2Container)
	require.NoError(t, err)

	out, err := yaml.Dump(c)
	require.NoError(t, err)
	require.NotContains(t, string(out), "version")
	require.NotContains(t, string(out), "instance")
	require.NotContains(t, string(out), "pools")
	require.NotContains(t, string(out), "volumes")
}
