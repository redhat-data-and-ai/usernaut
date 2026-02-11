package periodicjobs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ldapmocks "github.com/redhat-data-and-ai/usernaut/internal/controller/mocks"
	clientmocks "github.com/redhat-data-and-ai/usernaut/internal/controller/periodicjobs/mocks"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

// setupTestConfig creates a minimal test config environment to avoid file read panics
// Returns a cleanup function that should be called with defer
func setupTestConfig(t *testing.T) func() {
	tempDir, err := os.MkdirTemp("", "usernaut-test")
	require.NoError(t, err)

	configDir := filepath.Join(tempDir, "appconfig")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Create a minimal test config without file references
	configContent := `app:
  name: "usernaut-test"
  version: "0.0.1"
  environment: "test"
cache:
  driver: "memory"
  inmemory:
    defaultExpiration: -1
    cleanupInterval: -1
usernautUserOffboardingInterval: "24h"
offboardUserExclusionListConfigPath: "test_offboard_user_exclusion_list"
`
	err = os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Create exclusion list file (empty by default, tests can override if needed)
	exclusionListContent := `exclusions: []
`
	exclusionListFile := filepath.Join(configDir, "test_offboard_user_exclusion_list.yaml")
	err = os.WriteFile(exclusionListFile, []byte(exclusionListContent), 0644)
	require.NoError(t, err)

	// Set WORKDIR to point to our test directory
	oldWorkdir := os.Getenv("WORKDIR")
	err = os.Setenv("WORKDIR", tempDir)
	require.NoError(t, err)

	// Force reload of config to pick up the test config
	_, err = config.LoadConfig("default")
	require.NoError(t, err)

	return func() {
		if oldWorkdir != "" {
			_ = os.Setenv("WORKDIR", oldWorkdir)
		} else {
			_ = os.Unsetenv("WORKDIR")
		}
		_ = os.RemoveAll(tempDir)
	}
}

// TestUserOffboardingJob tests the offboarding job using mocks
func TestUserOffboardingJob(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockBackendClient := clientmocks.NewMockClient(ctrl)

	// Create in-memory cache (no need to mock, it's fast and reliable)
	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err, "Failed to create in-memory cache")

	// Create store layer
	dataStore := store.New(inMemCache)

	// Setup test data
	ctx := context.Background()
	testUser := &structs.User{
		ID:          "test_user_123",
		UserName:    "testuser",
		Email:       "testuser@example.com",
		FirstName:   "Test",
		LastName:    "User",
		DisplayName: "Test User",
		Role:        "test_role",
	}

	// Setup cache with user data using store layer
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_fivetran", testUser.ID)
	require.NoError(t, err)

	// Create backend clients map
	backendClients := map[string]clients.Client{
		"fivetran_fivetran": mockBackendClient,
	}

	// Create the job
	sharedCacheMutex := &sync.RWMutex{}
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("User_Not_In_LDAP_Should_Be_Offboarded", func(t *testing.T) {
		// Setup: LDAP returns ErrNoUserFound (user not found)
		// Note: getUserListFromCache returns emails, so LDAP is called with email using GetUserLDAPDataByEmail
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), testUser.Email).
			Return(nil, ldap.ErrNoUserFound).
			Times(1)

		// Setup: Backend client should be called to delete the user
		mockBackendClient.EXPECT().
			DeleteUser(gomock.Any(), testUser.ID).
			Return(nil).
			Times(1)

		// Run the job
		err := job.Run(ctx)
		assert.NoError(t, err)

		// Verify user is removed from cache using store layer
		exists, err := dataStore.User.Exists(ctx, testUser.Email)
		require.NoError(t, err)
		assert.False(t, exists, "User should be removed from cache")
	})

	// Reset cache for next test using store layer
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_fivetran", testUser.ID)
	require.NoError(t, err)

	t.Run("User_In_LDAP_Should_Not_Be_Offboarded", func(t *testing.T) {
		// Setup: LDAP returns user data (user found)
		// Note: getUserListFromCache returns emails, so LDAP is called with email using GetUserLDAPDataByEmail
		ldapData := map[string]interface{}{
			"mail": testUser.Email,
		}
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), testUser.Email).
			Return(ldapData, nil).
			Times(1)

		// Backend client should NOT be called to delete the user
		// (no EXPECT call means it should not be called)

		// Run the job
		err := job.Run(ctx)
		assert.NoError(t, err)

		// Verify user is still in cache using store layer
		exists, err := dataStore.User.Exists(ctx, testUser.Email)
		require.NoError(t, err)
		assert.True(t, exists, "User should remain in cache")
	})
}

// TestUserOffboardingJobBackendErrors tests error handling
func TestUserOffboardingJobBackendErrors(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockBackendClient := clientmocks.NewMockClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	// Create store layer
	dataStore := store.New(inMemCache)

	ctx := context.Background()
	testUser := &structs.User{
		ID:       "test_user_456",
		UserName: "erroruser",
		Email:    "erroruser@example.com",
	}

	// Setup cache using store layer
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_fivetran", testUser.ID)
	require.NoError(t, err)

	backendClients := map[string]clients.Client{
		"fivetran_fivetran": mockBackendClient,
	}

	sharedCacheMutex := &sync.RWMutex{}
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("Backend_Delete_Error_Should_Be_Logged", func(t *testing.T) {
		// LDAP says user doesn't exist
		// Note: getUserListFromCache returns emails, so LDAP is called with email using GetUserLDAPDataByEmail
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), testUser.Email).
			Return(nil, ldap.ErrNoUserFound).
			Times(1)

		// Backend delete fails
		mockBackendClient.EXPECT().
			DeleteUser(gomock.Any(), testUser.ID).
			Return(errors.New("backend service unavailable")).
			Times(1)

		// Run the job - should handle the error gracefully
		err := job.Run(ctx)
		// The job should return an error when backend deletion fails
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "backend service unavailable")
	})
}

// TestUserOffboardingJobEmptyUserList tests handling of empty user list
func TestUserOffboardingJobEmptyUserList(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	// Create store layer
	dataStore := store.New(inMemCache)

	ctx := context.Background()

	backendClients := map[string]clients.Client{}

	sharedCacheMutex := &sync.RWMutex{}
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	// No LDAP or backend calls should be made
	err = job.Run(ctx)
	assert.NoError(t, err)
}

// TestUserOffboardingJobMultipleBackends tests offboarding from multiple backends
func TestUserOffboardingJobMultipleBackends(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockFivetranClient := clientmocks.NewMockClient(ctrl)
	mockSnowflakeClient := clientmocks.NewMockClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	// Create store layer
	dataStore := store.New(inMemCache)

	ctx := context.Background()
	testUser := &structs.User{
		ID:       "multi_backend_user",
		UserName: "multiuser",
		Email:    "multiuser@example.com",
	}

	// Setup user with multiple backends using store layer
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_prod", "fivetran_id_123")
	require.NoError(t, err)
	err = dataStore.User.SetBackend(ctx, testUser.Email, "snowflake_prod", "snowflake_id_456")
	require.NoError(t, err)

	backendClients := map[string]clients.Client{
		"fivetran_prod":  mockFivetranClient,
		"snowflake_prod": mockSnowflakeClient,
	}

	sharedCacheMutex := &sync.RWMutex{}
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	// User not in LDAP
	// Note: getUserListFromCache returns emails, so LDAP is called with email using GetUserLDAPDataByEmail
	mockLDAPClient.EXPECT().
		GetUserLDAPDataByEmail(gomock.Any(), testUser.Email).
		Return(nil, ldap.ErrNoUserFound).
		Times(1)

	// Both backends should be called to delete the user
	mockFivetranClient.EXPECT().
		DeleteUser(gomock.Any(), "fivetran_id_123").
		Return(nil).
		Times(1)

	mockSnowflakeClient.EXPECT().
		DeleteUser(gomock.Any(), "snowflake_id_456").
		Return(nil).
		Times(1)

	// Run the job
	err = job.Run(ctx)
	assert.NoError(t, err)

	// Verify user is removed from cache using store layer
	exists, err := dataStore.User.Exists(ctx, testUser.Email)
	require.NoError(t, err)
	assert.False(t, exists, "User should be removed from cache")
}

// TestUserOffboardingJobInterval tests the GetInterval and GetName methods
func TestUserOffboardingJobInterval(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	dataStore := store.New(inMemCache)
	backendClients := map[string]clients.Client{}
	sharedCacheMutex := &sync.RWMutex{}

	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("GetName_Returns_Correct_Name", func(t *testing.T) {
		name := job.GetName()
		assert.Equal(t, UserOffboardingJobName, name, "GetName should return the correct job name")
	})

	t.Run("GetInterval_Returns_Valid_Duration", func(t *testing.T) {
		interval := job.GetInterval()
		// Should return at least the default interval (24 hours)
		assert.GreaterOrEqual(t, interval, DefaultUserOffboardingJobInterval,
			"GetInterval should return at least the default interval")
		// Should be a positive duration
		assert.Greater(t, interval, time.Duration(0),
			"GetInterval should return a positive duration")
	})

	t.Run("GetInterval_Returns_Consistent_Value", func(t *testing.T) {
		// GetInterval should return a consistent value (either default or configured)
		interval1 := job.GetInterval()
		interval2 := job.GetInterval()
		assert.Equal(t, interval1, interval2,
			"GetInterval should return a consistent value on multiple calls")
		// Should be at least the default interval
		assert.GreaterOrEqual(t, interval1, DefaultUserOffboardingJobInterval,
			"GetInterval should return at least the default interval")
	})
}

// TestUserOffboardingJobEmptyUserKey tests handling of empty user keys
func TestUserOffboardingJobEmptyUserKey(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockBackendClient := clientmocks.NewMockClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	dataStore := store.New(inMemCache)
	ctx := context.Background()

	// Setup cache with valid user and empty user key
	testUser := &structs.User{
		ID:       "test_user_123",
		Email:    "testuser@example.com",
		UserName: "testuser",
	}

	// Add valid user
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_fivetran", testUser.ID)
	require.NoError(t, err)

	// Add empty user key by directly setting an empty key in cache
	// This simulates a blank userKey in the cache
	// (key "user:" results in empty email after prefix removal)
	backendData := `{"fivetran_fivetran":"empty_user_id"}`
	err = inMemCache.Set(ctx, "user:", backendData, 0)
	require.NoError(t, err)

	backendClients := map[string]clients.Client{
		"fivetran_fivetran": mockBackendClient,
	}

	sharedCacheMutex := &sync.RWMutex{}
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("Empty_UserKey_Should_Be_Skipped", func(t *testing.T) {
		// Valid user should be processed
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), testUser.Email).
			Return(nil, ldap.ErrNoUserFound).
			Times(1)

		mockBackendClient.EXPECT().
			DeleteUser(gomock.Any(), testUser.ID).
			Return(nil).
			Times(1)

		// Empty user key should be skipped (no LDAP or backend calls for it)
		// The job should complete successfully despite the empty key

		err := job.Run(ctx)
		assert.NoError(t, err)

		// Valid user should be removed
		exists, err := dataStore.User.Exists(ctx, testUser.Email)
		require.NoError(t, err)
		assert.False(t, exists, "Valid user should be removed from cache")
	})
}

// TestUserOffboardingJobExclusionList tests handling of users in exclusion list
func TestUserOffboardingJobExclusionList(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockBackendClient := clientmocks.NewMockClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	dataStore := store.New(inMemCache)
	ctx := context.Background()

	// Setup test users
	excludedUser := &structs.User{
		ID:       "excluded_user_123",
		Email:    "excluded@example.com",
		UserName: "excluded",
	}

	normalUser := &structs.User{
		ID:       "normal_user_456",
		Email:    "normal@example.com",
		UserName: "normal",
	}

	// Add users to cache
	err = dataStore.User.SetBackend(ctx, excludedUser.Email, "fivetran_fivetran", excludedUser.ID)
	require.NoError(t, err)
	err = dataStore.User.SetBackend(ctx, normalUser.Email, "fivetran_fivetran", normalUser.ID)
	require.NoError(t, err)

	// Create exclusion list file with excluded user
	tempDir := os.Getenv("WORKDIR")
	configDir := filepath.Join(tempDir, "appconfig")
	exclusionListContent := `exclusions:
  - excluded@example.com
`
	exclusionListFile := filepath.Join(configDir, "test_offboard_user_exclusion_list.yaml")
	err = os.WriteFile(exclusionListFile, []byte(exclusionListContent), 0644)
	require.NoError(t, err)

	backendClients := map[string]clients.Client{
		"fivetran_fivetran": mockBackendClient,
	}

	sharedCacheMutex := &sync.RWMutex{}
	// Create job after exclusion list file is created so it loads the exclusion list
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("Excluded_User_Should_Be_Skipped", func(t *testing.T) {
		// Excluded user should NOT be checked in LDAP or deleted from backend
		// (no EXPECT calls for excluded user)

		// Normal user should be processed
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), normalUser.Email).
			Return(nil, ldap.ErrNoUserFound).
			Times(1)

		mockBackendClient.EXPECT().
			DeleteUser(gomock.Any(), normalUser.ID).
			Return(nil).
			Times(1)

		err := job.Run(ctx)
		assert.NoError(t, err)

		// Excluded user should still be in cache
		exists, err := dataStore.User.Exists(ctx, excludedUser.Email)
		require.NoError(t, err)
		assert.True(t, exists, "Excluded user should remain in cache")

		// Normal user should be removed
		exists, err = dataStore.User.Exists(ctx, normalUser.Email)
		require.NoError(t, err)
		assert.False(t, exists, "Normal user should be removed from cache")
	})
}

// TestUserOffboardingJobExclusionListFromURL tests loading exclusion list from HTTP URL
func TestUserOffboardingJobExclusionListFromURL(t *testing.T) {
	defer setupTestConfig(t)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLDAPClient := ldapmocks.NewMockLDAPClient(ctrl)
	mockBackendClient := clientmocks.NewMockClient(ctrl)

	cacheConfig := &inmemory.Config{
		DefaultExpiration: 60,
		CleanupInterval:   120,
	}
	inMemCache, err := inmemory.NewCache(cacheConfig)
	require.NoError(t, err)

	dataStore := store.New(inMemCache)
	ctx := context.Background()

	// Setup test users
	excludedUser := &structs.User{
		ID:       "excluded_user_123",
		Email:    "excluded@example.com",
		UserName: "excluded",
	}

	normalUser := &structs.User{
		ID:       "normal_user_456",
		Email:    "normal@example.com",
		UserName: "normal",
	}

	// Add users to cache
	err = dataStore.User.SetBackend(ctx, excludedUser.Email, "fivetran_fivetran", excludedUser.ID)
	require.NoError(t, err)
	err = dataStore.User.SetBackend(ctx, normalUser.Email, "fivetran_fivetran", normalUser.ID)
	require.NoError(t, err)

	// Create a test HTTP server that serves the exclusion list
	exclusionListContent := `exclusions:
  - excluded@example.com
`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/config/exclusion_list.yaml" {
			w.Header().Set("Content-Type", "application/yaml")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(exclusionListContent))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Update config to use HTTP URL
	tempDir := os.Getenv("WORKDIR")
	configDir := filepath.Join(tempDir, "appconfig")
	configContent := `app:
  name: "usernaut-test"
  version: "0.0.1"
  environment: "test"
cache:
  driver: "memory"
  inmemory:
    defaultExpiration: -1
    cleanupInterval: -1
usernautUserOffboardingInterval: "24h"
offboardUserExclusionListConfigPath: "` + server.URL + `/config/exclusion_list.yaml"
`
	err = os.WriteFile(filepath.Join(configDir, "default.yaml"), []byte(configContent), 0644)
	require.NoError(t, err)

	// Force reload of config
	_, err = config.LoadConfig("default")
	require.NoError(t, err)

	backendClients := map[string]clients.Client{
		"fivetran_fivetran": mockBackendClient,
	}

	sharedCacheMutex := &sync.RWMutex{}
	// Create job after config is updated so it loads the exclusion list from URL
	job := NewUserOffboardingJob(
		sharedCacheMutex,
		dataStore,
		mockLDAPClient,
		backendClients,
	)

	t.Run("Excluded_User_From_URL_Should_Be_Skipped", func(t *testing.T) {
		// Excluded user should NOT be checked in LDAP or deleted from backend
		// (no EXPECT calls for excluded user)

		// Normal user should be processed
		mockLDAPClient.EXPECT().
			GetUserLDAPDataByEmail(gomock.Any(), normalUser.Email).
			Return(nil, ldap.ErrNoUserFound).
			Times(1)

		mockBackendClient.EXPECT().
			DeleteUser(gomock.Any(), normalUser.ID).
			Return(nil).
			Times(1)

		err := job.Run(ctx)
		assert.NoError(t, err)

		// Excluded user should still be in cache
		exists, err := dataStore.User.Exists(ctx, excludedUser.Email)
		require.NoError(t, err)
		assert.True(t, exists, "Excluded user should remain in cache")

		// Normal user should be removed
		exists, err = dataStore.User.Exists(ctx, normalUser.Email)
		require.NoError(t, err)
		assert.False(t, exists, "Normal user should be removed from cache")
	})
}
