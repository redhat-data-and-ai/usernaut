package periodicjobs

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ldapmocks "github.com/redhat-data-and-ai/usernaut/internal/controller/mocks"
	clientmocks "github.com/redhat-data-and-ai/usernaut/internal/controller/periodicjobs/mocks"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

// TestUserOffboardingJob tests the offboarding job using mocks
func TestUserOffboardingJob(t *testing.T) {
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

	// Add user to user_list using store layer
	err = dataStore.Meta.SetUserList(ctx, []string{testUser.UserName})
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
		mockLDAPClient.EXPECT().
			GetUserLDAPData(gomock.Any(), testUser.UserName).
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

		// Verify user is removed from user_list using store layer
		updatedUserList, err := dataStore.Meta.GetUserList(ctx)
		require.NoError(t, err)
		assert.NotContains(t, updatedUserList, testUser.UserName, "User should be removed from user list")
	})

	// Reset cache for next test using store layer
	err = dataStore.User.SetBackend(ctx, testUser.Email, "fivetran_fivetran", testUser.ID)
	require.NoError(t, err)
	err = dataStore.Meta.SetUserList(ctx, []string{testUser.UserName})
	require.NoError(t, err)

	t.Run("User_In_LDAP_Should_Not_Be_Offboarded", func(t *testing.T) {
		// Setup: LDAP returns user data (user found)
		ldapData := map[string]interface{}{
			"mail": testUser.Email,
		}
		mockLDAPClient.EXPECT().
			GetUserLDAPData(gomock.Any(), testUser.UserName).
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

		// Verify user is still in user_list using store layer
		updatedUserList, err := dataStore.Meta.GetUserList(ctx)
		require.NoError(t, err)
		assert.Contains(t, updatedUserList, testUser.UserName, "User should remain in user list")
	})
}

// TestUserOffboardingJobBackendErrors tests error handling
func TestUserOffboardingJobBackendErrors(t *testing.T) {
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

	err = dataStore.Meta.SetUserList(ctx, []string{testUser.UserName})
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
		mockLDAPClient.EXPECT().
			GetUserLDAPData(gomock.Any(), testUser.UserName).
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

	// Setup empty user list using store layer
	err = dataStore.Meta.SetUserList(ctx, []string{})
	require.NoError(t, err)

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

	err = dataStore.Meta.SetUserList(ctx, []string{testUser.UserName})
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
	mockLDAPClient.EXPECT().
		GetUserLDAPData(gomock.Any(), testUser.UserName).
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
