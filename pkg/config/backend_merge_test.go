package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadEnvironmentBackends(t *testing.T) {
	for _, tc := range []struct {
		name         string
		production   string
		keepDefaults bool
		want         []Backend
	}{
		{
			name: "merge reordered backends and append new backend",
			production: `backends:
  - name: snowflake
    type: snowflake
    enabled: true
    connection:
      pat: snowflake-token
  - name: astro
    type: astro
    enabled: true
    connection:
      api_token: PRODUCTION_TOKEN
      organization_id: production-org
`,
			want: []Backend{
				{
					Name: "astro", Type: "astro", Enabled: true,
					Connection: map[string]interface{}{
						"api_token": "production-token", "organization_id": "production-org",
					},
				},
				{
					Name: "snowflake", Type: "snowflake", Enabled: true,
					Connection: map[string]interface{}{"pat": "snowflake-token"},
				},
			},
		},
		{
			name:         "empty list keeps default backends",
			production:   "backends: []\n",
			keepDefaults: true,
			want: []Backend{{
				Name: "astro", Type: "astro",
				Connection: map[string]interface{}{
					"api_token": "default-token", "organization_id": "default-org",
				},
			}},
		},
		{
			name:         "omitted list keeps default backends",
			production:   "app:\n  environment: production\n",
			keepDefaults: true,
			want: []Backend{{
				Name: "astro", Type: "astro",
				Connection: map[string]interface{}{
					"api_token": "default-token", "organization_id": "default-org",
				},
			}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			productionToken := filepath.Join(dir, "production-token")
			require.NoError(t, os.WriteFile(productionToken, []byte("production-token\n"), 0600))

			// Overriding the token must not read the missing default token file.
			defaultToken := "file|" + filepath.Join(dir, "missing-default-token")
			if tc.keepDefaults {
				defaultToken = "default-token"
			}
			defaults := fmt.Sprintf(`app:
  name: default-app
backends:
  - name: astro
    type: astro
    enabled: false
    connection:
      api_token: %q
      organization_id: default-org
`, defaultToken)
			production := strings.ReplaceAll(tc.production, "PRODUCTION_TOKEN", fmt.Sprintf("%q", "file|"+productionToken))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yaml"), []byte(defaults), 0600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "production.yaml"), []byte(production), 0600))

			var cfg AppConfig
			var err error
			require.NotPanics(t, func() {
				err = NewConfig(NewOptions("yaml", dir, "default")).Load("production", &cfg)
			})
			require.NoError(t, err)
			assert.Equal(t, tc.want, cfg.Backends)
			assert.Equal(t, "default-app", cfg.App.Name)
		})
	}
}

func TestBackendMergeReplacesMatchingIdentity(t *testing.T) {
	dir := t.TempDir()
	defaults := `backends:
  - name: primary
    type: astro
    enabled: true
    depends_on:
      name: original
      type: rover
    connection:
      api_token: default-token
      organization_id: default-org
  - name: untouched
    type: rover
    enabled: true
    connection:
      base_url: https://rover.example
`
	production := `backends:
  - name: secondary
    type: astro
    connection:
      api_token: secondary-token
  - name: primary
    type: snowflake
    connection:
      pat: snowflake-token
  - name: primary
    type: astro
    enabled: false
    depends_on:
      name: updated
    connection:
      api_token: ""
      base_url: https://astro.example
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yaml"), []byte(defaults), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "production.yaml"), []byte(production), 0600))
	var cfg AppConfig
	require.NoError(t, NewConfig(NewOptions("yaml", dir, "default")).Load("production", &cfg))
	assert.Equal(t, []Backend{
		{
			Name: "primary", Type: "astro", Enabled: false,
			DependsOn: Dependant{Name: "updated"},
			Connection: map[string]interface{}{
				"api_token": "", "base_url": "https://astro.example",
			},
		},
		{
			Name: "untouched", Type: "rover", Enabled: true,
			Connection: map[string]interface{}{"base_url": "https://rover.example"},
		},
		{
			Name: "secondary", Type: "astro",
			Connection: map[string]interface{}{"api_token": "secondary-token"},
		},
		{
			Name: "primary", Type: "snowflake",
			Connection: map[string]interface{}{"pat": "snowflake-token"},
		},
	}, cfg.Backends)
}

func TestMergeBackendsDoesNotReplaceInputEntries(t *testing.T) {
	defaults := []Backend{{Name: "astro", Type: "astro", Enabled: true}}
	environment := []Backend{{Name: "astro", Type: "astro", Enabled: false}}
	assert.Equal(t, environment, mergeBackends(defaults, environment))
	assert.True(t, defaults[0].Enabled, "replacing a result entry must not replace the input entry")
	assert.False(t, environment[0].Enabled)
}

func TestLoadBackendsKeepsLiteralKeys(t *testing.T) {
	dir := t.TempDir()
	defaults := `backends:
  - name: astro
    type: astro
    enabled: true
    connection:
      custom.key: original
      options:
        region: default-region
        timeout: 30
`
	environment := `backends:
  - name: astro
    type: astro
    connection:
      custom.key: updated
      options:
        timeout: 0
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "default.yaml"), []byte(defaults), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "production.yaml"), []byte(environment), 0600))
	loader := NewConfig(NewOptions("yaml", dir, "default"))
	var cfg AppConfig
	require.NoError(t, loader.Load("production", &cfg))
	require.Len(t, cfg.Backends, 1)
	assert.False(t, cfg.Backends[0].Enabled, "omitted enabled must not inherit from the replaced default")
	assert.Equal(t, map[string]interface{}{
		"custom.key": "updated",
		"options":    map[string]interface{}{"timeout": 0},
	}, cfg.Backends[0].Connection)

	// Reusing the loader must not retain environment overrides in its defaults.
	var defaultConfig AppConfig
	require.NoError(t, loader.Load("default", &defaultConfig))
	require.Len(t, defaultConfig.Backends, 1)
	assert.Equal(t, "original", defaultConfig.Backends[0].Connection["custom.key"])
	assert.Equal(t, map[string]interface{}{"region": "default-region", "timeout": 30},
		defaultConfig.Backends[0].Connection["options"])
}
