/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteStringFromFile(t *testing.T) {
	secretFile := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(secretFile, []byte("secret-value\n"), 0600))

	value, err := substituteString(FilePrefix + secretFile)

	require.NoError(t, err)
	assert.Equal(t, "secret-value", value)
}

func TestSubstituteStringFromEnv(t *testing.T) {
	t.Setenv("CONFIG_SECRET", "env-value")

	value, err := substituteString(EnvPrefix + "CONFIG_SECRET")

	require.NoError(t, err)
	assert.Equal(t, "env-value", value)
}

func TestSubstituteStringPlainString(t *testing.T) {
	value, err := substituteString("plain-value")

	require.NoError(t, err)
	assert.Equal(t, "plain-value", value)
}

func TestSubstituteStringMissingFileReturnsError(t *testing.T) {
	missingFile := filepath.Join(t.TempDir(), "missing.txt")

	var err error
	require.NotPanics(t, func() {
		_, err = substituteString(FilePrefix + missingFile)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file "+missingFile)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestLoadReturnsFileSubstitutionError(t *testing.T) {
	configDir := t.TempDir()
	missingFile := filepath.Join(configDir, "missing.txt")
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte("secret: default\n"), 0600))
	require.NoError(t, os.WriteFile(
		filepath.Join(configDir, "dev.yaml"),
		[]byte("secret: \"file|"+missingFile+"\"\n"),
		0600,
	))

	var loaded struct {
		Secret string
	}
	var err error
	require.NotPanics(t, func() {
		err = NewConfig(NewOptions("yaml", configDir, "default")).Load("dev", &loaded)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file "+missingFile)
	assert.ErrorIs(t, err, os.ErrNotExist)
}
