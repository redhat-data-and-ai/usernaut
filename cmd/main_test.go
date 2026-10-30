package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	k8sCache "sigs.k8s.io/controller-runtime/pkg/cache"
)

func TestParseWatchedNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected map[string]k8sCache.Config
	}{
		{
			name:     "empty value falls back to the default namespace",
			value:    "",
			expected: map[string]k8sCache.Config{defaultWatchedNamespace: {}},
		},
		{
			name:     "only separators falls back to the default namespace",
			value:    " , ,",
			expected: map[string]k8sCache.Config{defaultWatchedNamespace: {}},
		},
		{
			name:     "single namespace",
			value:    "usernaut",
			expected: map[string]k8sCache.Config{"usernaut": {}},
		},
		{
			name:  "comma separated namespaces are trimmed and deduplicated",
			value: " usernaut , dataverse ,,usernaut, tenant-a ",
			expected: map[string]k8sCache.Config{
				"usernaut":  {},
				"dataverse": {},
				"tenant-a":  {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, parseWatchedNamespaces(tt.value))
		})
	}
}
