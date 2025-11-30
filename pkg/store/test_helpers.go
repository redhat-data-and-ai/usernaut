package store

import (
	"context"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EntityStoreInterface defines common operations for both UserStore and TeamStore
// This is used for test helpers to avoid duplication
type EntityStoreInterface interface {
	GetBackends(ctx context.Context, identifier string) (map[string]string, error)
	SetBackend(ctx context.Context, identifier, backendKey, backendID string) error
	DeleteBackend(ctx context.Context, identifier, backendKey string) error
	Delete(ctx context.Context, identifier string) error
	Exists(ctx context.Context, identifier string) (bool, error)
}

// SetBackendTestCase defines a test case for SetBackend operations
type SetBackendTestCase struct {
	Name        string
	Identifier  string // email for users, name for teams
	BackendKey  string
	BackendID   string
	SetupFunc   func(t *testing.T, store EntityStoreInterface)
	VerifyFunc  func(t *testing.T, store EntityStoreInterface)
	WantErr     bool
	ErrContains string
}

// DeleteBackendTestCase defines a test case for DeleteBackend operations
type DeleteBackendTestCase struct {
	Name        string
	Identifier  string
	BackendKey  string
	SetupFunc   func(t *testing.T, store EntityStoreInterface)
	VerifyFunc  func(t *testing.T, store EntityStoreInterface)
	WantErr     bool
	ErrContains string
}

// DeleteTestCase defines a test case for Delete operations
type DeleteTestCase struct {
	Name       string
	Identifier string
	SetupFunc  func(t *testing.T, store EntityStoreInterface)
	WantErr    bool
}

// ExistsTestCase defines a test case for Exists operations
type ExistsTestCase struct {
	Name       string
	Identifier string
	SetupFunc  func(t *testing.T, store EntityStoreInterface)
	WantExist  bool
}

// RunSetBackendTests runs table-driven tests for SetBackend operation
func RunSetBackendTests(t *testing.T, tests []SetBackendTestCase, storeFactory func() EntityStoreInterface) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			store := storeFactory()
			tt.SetupFunc(t, store)

			err := store.SetBackend(context.Background(), tt.Identifier, tt.BackendKey, tt.BackendID)

			if tt.WantErr {
				assert.Error(t, err)
				if tt.ErrContains != "" {
					assert.Contains(t, err.Error(), tt.ErrContains)
				}
			} else {
				assert.NoError(t, err)
				tt.VerifyFunc(t, store)
			}
		})
	}
}

// RunDeleteBackendTests runs table-driven tests for DeleteBackend operation
func RunDeleteBackendTests(t *testing.T, tests []DeleteBackendTestCase, storeFactory func() EntityStoreInterface) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			store := storeFactory()
			tt.SetupFunc(t, store)

			err := store.DeleteBackend(context.Background(), tt.Identifier, tt.BackendKey)

			if tt.WantErr {
				assert.Error(t, err)
				if tt.ErrContains != "" {
					assert.Contains(t, err.Error(), tt.ErrContains)
				}
			} else {
				assert.NoError(t, err)
				tt.VerifyFunc(t, store)
			}
		})
	}
}

// RunDeleteTests runs table-driven tests for Delete operation
func RunDeleteTests(t *testing.T, tests []DeleteTestCase, storeFactory func() EntityStoreInterface) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			store := storeFactory()
			tt.SetupFunc(t, store)

			err := store.Delete(context.Background(), tt.Identifier)

			if tt.WantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify deletion
				exists, err := store.Exists(context.Background(), tt.Identifier)
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}

// RunExistsTests runs table-driven tests for Exists operation
func RunExistsTests(t *testing.T, tests []ExistsTestCase, storeFactory func() EntityStoreInterface) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			store := storeFactory()
			tt.SetupFunc(t, store)

			exists, err := store.Exists(context.Background(), tt.Identifier)

			assert.NoError(t, err)
			assert.Equal(t, tt.WantExist, exists)
		})
	}
}

// GetBackendsTestCase defines a test case for GetBackends operations
type GetBackendsTestCase struct {
	Name        string
	Identifier  string
	SetupFunc   func(t *testing.T, store EntityStoreInterface, c cache.Cache)
	Want        map[string]string
	WantErr     bool
	ErrContains string
}

// RunGetBackendsTests runs table-driven tests for GetBackends operation
func RunGetBackendsTests(
	t *testing.T,
	tests []GetBackendsTestCase,
	storeFactory func() (EntityStoreInterface, cache.Cache),
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			store, c := storeFactory()
			tt.SetupFunc(t, store, c)

			got, err := store.GetBackends(context.Background(), tt.Identifier)

			if tt.WantErr {
				assert.Error(t, err)
				if tt.ErrContains != "" {
					assert.Contains(t, err.Error(), tt.ErrContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.Want, got)
			}
		})
	}
}
