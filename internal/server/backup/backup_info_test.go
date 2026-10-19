package backup

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func tarWithIndex(t *testing.T, index string) *os.File {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := tw.WriteHeader(&tar.Header{Name: backupIndexPath, Mode: 0o644, Size: int64(len(index))})
	require.NoError(t, err)

	_, err = tw.Write([]byte(index))
	require.NoError(t, err)
	require.NoError(t, tw.Close())

	path := filepath.Join(t.TempDir(), "backup.tar")
	require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o600))

	f, err := os.Open(path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = f.Close() })

	return f
}

func TestGetInfoLXDv2(t *testing.T) {
	f := tarWithIndex(t, `name: c1
backend: dir
pool: default
type: container
config:
  version: 2
  instance:
    name: c1
    type: container
  pools:
  - name: default
  volumes:
  - name: c1
    type: container
    snapshots:
    - name: snap0
`)

	info, err := GetInfo(f, nil, f.Name())
	require.NoError(t, err)
	require.NotNil(t, info.Config.Container)
	require.Equal(t, "c1", info.Config.Container.Name)
	require.Equal(t, "default", info.Config.Pool.Name)
	require.Len(t, info.Config.VolumeSnapshots, 1)
}

func TestGetInfoV1(t *testing.T) {
	f := tarWithIndex(t, `name: c1
backend: dir
pool: default
config:
  container:
    name: c1
`)

	info, err := GetInfo(f, nil, f.Name())
	require.NoError(t, err)
	require.Equal(t, "c1", info.Config.Container.Name)
}

func TestGetInfoUnsupportedVersion(t *testing.T) {
	f := tarWithIndex(t, "name: c1\nconfig:\n  version: 3\n")

	_, err := GetInfo(f, nil, f.Name())
	require.ErrorContains(t, err, "Unsupported backup metadata version")
}

func TestParseConfigYamlFileLXDv2(t *testing.T) {
	path := filepath.Join(t.TempDir(), "backup.yaml")
	require.NoError(t, os.WriteFile(path, []byte("version: 2\ninstance:\n  name: c1\npools:\n- name: default\nvolumes:\n- name: c1\n"), 0o600))

	c, err := ParseConfigYamlFile(path)
	require.NoError(t, err)
	require.Equal(t, "container", c.Container.Type)
	require.Equal(t, "c1", c.Volume.Name)
}
