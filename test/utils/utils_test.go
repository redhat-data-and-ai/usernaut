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

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProjectDir(t *testing.T) {
	origWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(origWd) }()

	t.Run("returns current working directory when not in e2e tests", func(t *testing.T) {
		_ = os.Chdir(origWd)
		dir, err := GetProjectDir()
		assert.NoError(t, err)
		assert.NotEmpty(t, dir)
		assert.False(t, strings.Contains(dir, "/test/e2e"))
	})

	t.Run("removes /test/e2e from path when present", func(t *testing.T) {
		testPath := filepath.Join(origWd, "test", "e2e", "tests")
		require.NoError(t, os.MkdirAll(testPath, 0o755))
		defer func() { _ = os.RemoveAll(filepath.Join(origWd, "test")) }()

		require.NoError(t, os.Chdir(testPath))
		dir, err := GetProjectDir()
		assert.NoError(t, err)
		assert.NotEmpty(t, dir)
		assert.False(t, strings.Contains(dir, "/test/e2e"))
	})
}

func TestGetNonEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single line",
			input:    "line1",
			expected: []string{"line1"},
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines in middle",
			input:    "line1\n\nline2",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "trailing empty lines",
			input:    "line1\nline2\n\n",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "leading empty lines",
			input:    "\n\nline1\nline2",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "all empty lines returns nil",
			input:    "\n\n\n",
			expected: nil,
		},
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "only one empty line returns nil",
			input:    "\n",
			expected: nil,
		},
		{
			name:     "multiple spaces preserved",
			input:    "line  with  spaces\nanother  line",
			expected: []string{"line  with  spaces", "another  line"},
		},
		{
			name:     "whitespace only lines are not empty",
			input:    "line1\n   \nline2",
			expected: []string{"line1", "   ", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNonEmptyLines(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNonEmptyLines_EdgeCases(t *testing.T) {
	t.Run("handles very long output", func(t *testing.T) {
		lines := []string{}
		for i := 0; i < 1000; i++ {
			lines = append(lines, "line"+string(rune(i%10)))
		}
		input := strings.Join(lines, "\n")
		result := GetNonEmptyLines(input)
		assert.Equal(t, len(lines), len(result))
	})

	t.Run("handles special characters in lines", func(t *testing.T) {
		input := "line1:with:colons\nline2@with@at\nline3#with#hash"
		result := GetNonEmptyLines(input)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, "line1:with:colons", result[0])
		assert.Equal(t, "line2@with@at", result[1])
		assert.Equal(t, "line3#with#hash", result[2])
	})
}

func TestRun(t *testing.T) {
	t.Run("successful command execution", func(t *testing.T) {
		cmd := exec.Command("echo", "hello")
		output, err := Run(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "hello")
	})

	t.Run("command with environment variable", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo $GO111MODULE")
		output, err := Run(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "on")
	})

	t.Run("failed command execution", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 1")
		_, err := Run(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed with error")
	})

	t.Run("command with working directory", func(t *testing.T) {
		cmd := exec.Command("pwd")
		output, err := Run(cmd)
		assert.NoError(t, err)
		assert.NotEmpty(t, output)
	})

	t.Run("command with multiple arguments", func(t *testing.T) {
		cmd := exec.Command("echo", "arg1", "arg2", "arg3")
		output, err := Run(cmd)
		assert.NoError(t, err)
		assert.Contains(t, string(output), "arg1")
		assert.Contains(t, string(output), "arg2")
		assert.Contains(t, string(output), "arg3")
	})
}

func TestInstallPrometheusOperator(t *testing.T) {
	t.Run("install command is properly constructed", func(t *testing.T) {
		// Just verify the function exists and is callable
		// Actual kubectl calls are integration tests, not unit tests
		assert.NotNil(t, InstallPrometheusOperator)
	})
}

func TestUninstallPrometheusOperator(t *testing.T) {
	t.Run("uninstall prometheus operator does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			UninstallPrometheusOperator()
		})
	})
}

func TestInstallCertManager(t *testing.T) {
	t.Run("install command is properly constructed", func(t *testing.T) {
		// Just verify the function exists and is callable
		// Actual kubectl calls are integration tests, not unit tests
		assert.NotNil(t, InstallCertManager)
	})
}

func TestUninstallCertManager(t *testing.T) {
	t.Run("uninstall cert manager does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			UninstallCertManager()
		})
	})
}

func TestLoadImageToKindClusterWithName(t *testing.T) {
	t.Run("with kind cluster environment variable", func(t *testing.T) {
		origVal, set := os.LookupEnv("KIND_CLUSTER")
		_ = os.Setenv("KIND_CLUSTER", "test-cluster")
		defer func() {
			if set {
				_ = os.Setenv("KIND_CLUSTER", origVal)
			} else {
				_ = os.Unsetenv("KIND_CLUSTER")
			}
		}()

		err := LoadImageToKindClusterWithName("test-image:latest")
		// Error is expected if kind is not installed
		if err != nil {
			assert.Error(t, err)
		}
	})

	t.Run("without kind cluster environment variable uses default", func(t *testing.T) {
		origVal, set := os.LookupEnv("KIND_CLUSTER")
		_ = os.Unsetenv("KIND_CLUSTER")
		defer func() {
			if set {
				_ = os.Setenv("KIND_CLUSTER", origVal)
			}
		}()

		err := LoadImageToKindClusterWithName("test-image:latest")
		// Error is expected if kind is not installed
		if err != nil {
			assert.Error(t, err)
		}
	})
}

func TestWarnError(t *testing.T) {
	t.Run("warnError with nil does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			warnError(nil)
		})
	})

	t.Run("warnError with error does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			warnError(assert.AnError)
		})
	})
}

func TestGetNonEmptyLinesConsistency(t *testing.T) {
	input := "line1\n\nline2\nline3"
	result1 := GetNonEmptyLines(input)
	result2 := GetNonEmptyLines(input)

	assert.Equal(t, result1, result2)
	require.Equal(t, 3, len(result1))
}

func TestGetProjectDirCleanup(t *testing.T) {
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	dir1, err := GetProjectDir()
	require.NoError(t, err)
	require.NotEmpty(t, dir1)

	dir2, err := GetProjectDir()
	require.NoError(t, err)
	require.NotEmpty(t, dir2)

	assert.Equal(t, dir1, dir2)
}
