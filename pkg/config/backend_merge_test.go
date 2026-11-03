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
			name: "replace default backends",
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
					Name: "snowflake", Type: "snowflake", Enabled: true,
					Connection: map[string]interface{}{"pat": "snowflake-token"},
				},
				{
					Name: "astro", Type: "astro", Enabled: true,
					Connection: map[string]interface{}{
						"api_token": "production-token", "organization_id": "production-org",
					},
				},
			},
		},
		{
			name:       "empty list removes default backends",
			production: "backends: []\n",
			want:       []Backend{},
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

			// Replacing or removing defaults must not read their missing token file.
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
