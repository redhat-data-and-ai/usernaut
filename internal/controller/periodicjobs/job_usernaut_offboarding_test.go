package periodicjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

// UserOffboardingJobTestSuite defines the test suite for UserOffboardingJob integration tests
type UserOffboardingJobTestSuite struct {
	suite.Suite
	ctx            context.Context
	cacheClient    cache.Cache
	ldapClient     ldap.LDAPClient
	fivetranClient clients.Client
	job            *UserOffboardingJob
	vinodUser      *structs.User
	testUserKey    string
}

// SetupSuite runs once before all tests in the suite
func (suite *UserOffboardingJobTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Set APP_ENV to local to use local.yaml configuration
	_ = os.Setenv("APP_ENV", "local")

	// Initialize real Redis cache client using config from local.yaml
	appConfig, err := config.GetConfig()
	if err != nil {
		suite.T().Skipf("Failed to get config: %v. Ensure config is accessible.", err)
		return
	}
	require.NoError(suite.T(), err, "Failed to get config")

	// Use Redis cache from config, but override database to avoid conflicts with production
	cacheConfig := appConfig.Cache
	if cacheConfig.Redis != nil {
		// Use database 1 for tests to avoid conflicts with production (which uses database 0)
		cacheConfig.Redis.Database = 1
	}

	suite.cacheClient, err = cache.New(&cacheConfig)
	require.NoError(suite.T(), err, "Failed to initialize Redis cache client")

	// Initialize real LDAP client using config from local.yaml
	ldapConfig := appConfig.LDAP

	suite.ldapClient, err = ldap.InitLdap(ldapConfig)
	if err != nil {
		suite.T().Skipf("Failed to initialize LDAP client: %v. Ensure LDAP server is accessible.", err)
		return
	}

	fivetranAPIKey := fmt.Sprint(appConfig.Backends[0].Connection["apikey"])
	fivetranAPISecret := fmt.Sprint(appConfig.Backends[0].Connection["apisecret"])

	if fivetranAPIKey == "" || fivetranAPISecret == "" {
		suite.T().Skip("Fivetran API credentials not provided.")
		return
	}

	suite.fivetranClient = fivetran.NewClient(fivetranAPIKey, fivetranAPISecret)
}

// SetupTest runs before each test
func (suite *UserOffboardingJobTestSuite) SetupTest() {
	// Create Vinod user data for testing
	suite.vinodUser = &structs.User{
		ID:          "test_vinod_" + fmt.Sprintf("%d", time.Now().Unix()), // Unique ID for test isolation
		UserName:    "test_vinod",
		Email:       "test_vinod_" + fmt.Sprintf("%d", time.Now().Unix()) + "@example.com",
		FirstName:   "Vinod",
		LastName:    "Kumar1",
		DisplayName: "Vinod Kumar1",
		Role:        "mygroup_group",
	}

	suite.testUserKey = suite.vinodUser.UserName

	// Create UserOffboardingJob with real dependencies
	suite.job = &UserOffboardingJob{
		cacheClient: suite.cacheClient,
		ldapClient:  suite.ldapClient,
		backendClients: map[string]clients.Client{
			"fivetran_fivetran": suite.fivetranClient,
		},
		cacheMutex: sync.RWMutex{},
	}
}

// TearDownTest runs after each test
func (suite *UserOffboardingJobTestSuite) TearDownTest() {
	// Safety cleanup: Remove test user from cache if it still exists
	// (This is redundant for successful offboarding tests, but necessary for failed tests)
	if suite.vinodUser != nil {
		_ = suite.cacheClient.Delete(suite.ctx, suite.vinodUser.Email)
	}

	// Safety cleanup: Try to remove test user from Fivetran if it still exists
	// (This is redundant for successful offboarding tests, but necessary for failed tests)
	// Note: This is best effort cleanup, errors are ignored
	if suite.fivetranClient != nil && suite.vinodUser != nil {
		_ = suite.fivetranClient.DeleteUser(suite.ctx, suite.vinodUser.ID)
	}

	// Safety cleanup: Remove test user from user_list cache if it still exists
	// (This is redundant for successful offboarding tests, but necessary for failed tests)
	userListData, err := suite.cacheClient.Get(suite.ctx, "user_list")
	if err == nil {
		var userList []string
		if userListStr, ok := userListData.(string); ok {
			if json.Unmarshal([]byte(userListStr), &userList) == nil {
				// Remove test user from list (compare with full cache key)
				updatedList := make([]string, 0)
				for _, user := range userList {
					if user != suite.testUserKey {
						updatedList = append(updatedList, user)
					}
				}
				if updatedListJSON, err := json.Marshal(updatedList); err == nil {
					_ = suite.cacheClient.Set(suite.ctx, "user_list", string(updatedListJSON), cache.NoExpiration)
				}
			}
		}
	}
}

// TestCompleteOffboardingFlow tests the main offboarding scenario
func (suite *UserOffboardingJobTestSuite) TestCompleteOffboardingFlow() {
	// Step 1: Create Vinod in Fivetran backend
	suite.T().Log("Creating Vinod in Fivetran backend")
	createdUser, err := suite.fivetranClient.CreateUser(suite.ctx, suite.vinodUser)

	require.NoError(suite.T(), err, "Failed to create user in Fivetran")
	require.NotNil(suite.T(), createdUser, "Created user should not be nil")
	assert.Equal(suite.T(), suite.vinodUser.Email, createdUser.Email, "Email should match")

	// Update vinodUser with the ID returned by Fivetran
	suite.vinodUser.ID = createdUser.ID

	// Step 2: Add Vinod to Redis cache with backend mappings format
	suite.T().Log("Adding Vinod to Redis cache")
	// Create backend mappings in the new format: {"fivetran_fivetran": "user_id"}
	backendMappings := map[string]string{
		"fivetran_fivetran": suite.vinodUser.ID,
	}
	vinodUserJSON, err := json.Marshal(backendMappings)
	require.NoError(suite.T(), err, "Failed to marshal backend mappings")

	err = suite.cacheClient.Set(suite.ctx, suite.vinodUser.Email, string(vinodUserJSON), cache.NoExpiration)
	require.NoError(suite.T(), err, "Failed to set user in cache")

	// Add Vinod to user_list cache (using full cache key as expected by the job)
	userList := []string{suite.testUserKey} // Use full cache key, not just username
	userListJSON, err := json.Marshal(userList)
	require.NoError(suite.T(), err, "Failed to marshal user list")

	err = suite.cacheClient.Set(suite.ctx, "user_list", string(userListJSON), cache.NoExpiration)
	require.NoError(suite.T(), err, "Failed to set user list in cache")

	// Verify Vinod is in cache
	cachedData, err := suite.cacheClient.Get(suite.ctx, suite.vinodUser.Email)
	require.NoError(suite.T(), err, "Failed to get user from cache")
	assert.NotNil(suite.T(), cachedData, "Cached data should not be nil")

	// Step 3: Verify Vinod exists in Fivetran
	suite.T().Log("Verifying Vinod exists in Fivetran")
	fetchedUser, err := suite.fivetranClient.FetchUserDetails(suite.ctx, suite.vinodUser.ID)
	require.NoError(suite.T(), err, "Failed to fetch user from Fivetran")
	require.NotNil(suite.T(), fetchedUser, "Fetched user should not be nil")
	assert.Equal(suite.T(), suite.vinodUser.Email, fetchedUser.Email, "Email should match")

	// Step 4: First, let's check what LDAP returns for our test user
	suite.T().Log("Checking LDAP status for test user")
	ldapData, ldapErr := suite.ldapClient.GetUserLDAPData(suite.ctx, suite.vinodUser.UserName)
	suite.T().Logf("LDAP lookup result for %s: data=%v, error=%v", suite.vinodUser.UserName, ldapData, ldapErr)

	// Step 5: Debug - check what's in user_list
	userListData, _ := suite.cacheClient.Get(suite.ctx, "user_list")
	suite.T().Logf("user_list contents: %v", userListData)

	// Step 6: Run the periodic job
	suite.T().Log("Running the user offboarding job")
	err = suite.job.Run(suite.ctx)

	// The job might return an error if there are issues, but we expect it to process
	// We'll verify the actual results rather than just checking for no error
	suite.T().Logf("Job run result: %v", err)

	// Step 6: Verify results based on LDAP status
	if ldapErr != nil {
		// User not found in LDAP - should be offboarded
		suite.T().Log("User not found in LDAP - verifying offboarding")

		// Verify Vinod is removed from cache
		suite.T().Log("Verifying Vinod is removed from cache")
		cachedData, err := suite.cacheClient.Get(suite.ctx, suite.vinodUser.Email)
		assert.Error(suite.T(), err, "User should be removed from cache")
		// cachedData might be nil or empty string depending on cache implementation
		if cachedData != nil {
			assert.Empty(suite.T(), cachedData, "Cached data should be empty after removal")
		}

		// Verify Vinod is removed from user_list
		suite.T().Log("Verifying Vinod is removed from user_list")
		userListData, err := suite.cacheClient.Get(suite.ctx, "user_list")
		if err == nil {
			var updatedUserList []string
			userListStr, ok := userListData.(string)
			if ok {
				err = json.Unmarshal([]byte(userListStr), &updatedUserList)
				if err == nil {
					assert.NotContains(suite.T(), updatedUserList, suite.testUserKey, "User should be removed from user list")
				}
			}
		}

		// Verify Vinod is removed from Fivetran
		suite.T().Log("Verifying Vinod is removed from Fivetran")
		_, err = suite.fivetranClient.FetchUserDetails(suite.ctx, suite.vinodUser.ID)
		assert.Error(suite.T(), err, "User should be removed from Fivetran")
	} else {
		// User found in LDAP - should NOT be offboarded
		suite.T().Log("User found in LDAP - verifying user is preserved")

		// Verify Vinod is still in cache
		suite.T().Log("Verifying Vinod is still in cache")
		_, err = suite.cacheClient.Get(suite.ctx, suite.vinodUser.Email)
		assert.NoError(suite.T(), err, "User should remain in cache")

		// Verify Vinod is still in user_list
		suite.T().Log("Verifying Vinod is still in user_list")
		userListData, err := suite.cacheClient.Get(suite.ctx, "user_list")
		if err == nil {
			var currentUserList []string
			userListStr, ok := userListData.(string)
			if ok {
				err = json.Unmarshal([]byte(userListStr), &currentUserList)
				if err == nil {
					assert.Contains(suite.T(), currentUserList, suite.testUserKey, "User should remain in user list")
				}
			}
		}

		// Verify Vinod is still in Fivetran
		suite.T().Log("Verifying Vinod is still in Fivetran")
		_, err = suite.fivetranClient.FetchUserDetails(suite.ctx, suite.vinodUser.ID)
		assert.NoError(suite.T(), err, "User should remain in Fivetran")
	}
}

// TestUserOffboardingJob runs the test suite
func TestUserOffboardingJob(t *testing.T) {
	suite.Run(t, new(UserOffboardingJobTestSuite))
}
