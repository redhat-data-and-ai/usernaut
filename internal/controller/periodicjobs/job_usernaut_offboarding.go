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

// Package periodicjobs provides scheduled background jobs for the usernaut controller.
//
// This file implements the user offboarding periodic job that automatically removes
// inactive users from all backend systems when they are no longer found in LDAP.
package periodicjobs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// UserOffboardingJobName is the unique identifier for the user offboarding periodic job.
	UserOffboardingJobName = "usernaut_user_offboarding"

	// UserOffboardingJobInterval defines how often the user offboarding job runs.
	// Set to 24 hours to perform daily cleanup of inactive users.
	UserOffboardingJobInterval = 24 * time.Hour

	// UserCacheKeyPrefix is the Redis key prefix used to identify user entries in the cache.
	// All user keys should follow the pattern "user:{userID}".
	UserCacheKeyPrefix = "user:"
)

// UserOffboardingJob implements a periodic job that monitors user activity and automatically
// offboards inactive users from all configured backends.
//
// The job performs the following operations:
//  1. Scans Redis cache for all user entries
//  2. Verifies each user's status in LDAP directory
//  3. Offboards users who are no longer active in LDAP from all backends
//  4. Removes inactive users from the cache
//
// This ensures that user access is automatically revoked when users leave the organization
// or become inactive in the LDAP directory.
type UserOffboardingJob struct {

	// cacheClient provides access to the Redis cache containing user data.
	cacheClient cache.Cache

	// ldapClient enables verification of user status in the LDAP directory.
	ldapClient ldap.LDAPClient

	// backendClients contains all configured backend clients (Fivetran, Rover, etc.)
	// mapped by their unique identifier "{name}_{type}".
	backendClients map[string]clients.Client

	// cacheMutex prevents concurrent access to the cache during user offboarding operations.
	// This ensures that multiple reconcile loops don't interfere with each other when
	// reading or modifying user data in Redis.
	cacheMutex sync.RWMutex
}

// NewUserOffboardingJob creates and initializes a new UserOffboardingJob instance.
//
// This constructor:
//   - Loads the application configuration
//   - Initializes cache and LDAP clients
//   - Sets up all enabled backend clients
//   - Returns a fully configured job ready for execution
//
// Parameters:
//   - snowflakeEnvironment: The Snowflake environment identifier for this job instance
//
// Returns:
//   - *UserOffboardingJob: A configured job instance
//   - error: Any initialization error encountered
func NewUserOffboardingJob() (*UserOffboardingJob, error) {
	// Get application configuration
	appConfig, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	// Initialize cache client
	cacheClient, err := cache.New(&appConfig.Cache)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Initialize LDAP client
	ldapClient, err := ldap.InitLdap(appConfig.LDAP)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LDAP client: %w", err)
	}

	// Initialize backend clients
	backendClients := make(map[string]clients.Client)
	for _, backend := range appConfig.Backends {
		if backend.Enabled {
			client, err := clients.New(backend.Name, backend.Type, appConfig.BackendMap)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize backend client %s/%s: %w",
					backend.Type, backend.Name, err)
			}
			backendClients[fmt.Sprintf("%s_%s", backend.Name, backend.Type)] = client
		}
	}

	return &UserOffboardingJob{
		cacheClient:    cacheClient,
		ldapClient:     ldapClient,
		backendClients: backendClients,
	}, nil
}

// AddToPeriodicTaskManager registers this job with the provided periodic task manager.
//
// This method integrates the user offboarding job into the controller's periodic
// task execution system, allowing it to run at the configured interval.
//
// Parameters:
//   - mgr: The PeriodicTaskManager instance to register this job with
func (uoj *UserOffboardingJob) AddToPeriodicTaskManager(mgr *PeriodicTaskManager) {
	mgr.AddTask(uoj)
}

// GetInterval returns the execution interval for this periodic job.
//
// This method is required by the PeriodicTask interface and defines how often
// the user offboarding job should be executed.
//
// Returns:
//   - time.Duration: The interval between job executions (24 hours)
func (uoj *UserOffboardingJob) GetInterval() time.Duration {
	return UserOffboardingJobInterval
}

// GetName returns the unique name identifier for this periodic job.
//
// This method is required by the PeriodicTask interface and provides a
// human-readable name for logging and monitoring purposes.
//
// Returns:
//   - string: The job name "usernaut_user_offboarding"
func (uoj *UserOffboardingJob) GetName() string {
	return UserOffboardingJobName
}

// Run executes the main user offboarding logic.
//
// This method is required by the PeriodicTask interface and contains the core
// business logic for identifying and offboarding inactive users.
//
// The execution flow:
//  1. Retrieves all user keys from the cache
//  2. Processes each user to check LDAP status
//  3. Offboards users who are inactive in LDAP
//  4. Reports execution results and any errors
//
// Parameters:
//   - ctx: Context for cancellation and logging
//
// Returns:
//   - error: Any fatal error that occurred during execution, or a summary
//     of non-fatal errors if any users failed to process
func (uoj *UserOffboardingJob) Run(ctx context.Context) error {
	logger := log.FromContext(ctx)
	logger.Info("Starting user offboarding job")

	userKeys, err := uoj.getUserKeysFromCache(ctx)
	if err != nil {
		logger.Error(err, "Failed to get user keys from cache")
		return err
	}

	logger.Info("Found users in cache", "count", len(userKeys))

	result := uoj.processUsers(ctx, userKeys)

	logger.Info("User offboarding job completed",
		"totalUsers", len(userKeys),
		"offboardedUsers", result.offboardedCount,
		"errors", len(result.errors))

	if len(result.errors) > 0 {
		return fmt.Errorf("user offboarding completed with %d errors: %v", len(result.errors), result.errors)
	}

	return nil
}

// processingResult holds the results of processing multiple users during a job execution.
type processingResult struct {
	// offboardedCount tracks the number of users successfully offboarded
	offboardedCount int
	// errors contains all error messages encountered during processing
	errors []string
}

// processUsers iterates through all provided user keys and processes each user.
//
// This method coordinates the processing of multiple users, collecting results
// and errors from individual user processing operations.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userKeys: Slice of Redis keys identifying users to process
//
// Returns:
//   - processingResult: Summary of processing results including counts and errors
func (uoj *UserOffboardingJob) processUsers(ctx context.Context, userKeys []string) processingResult {
	var result processingResult

	for _, userKey := range userKeys {
		userID := strings.TrimPrefix(userKey, UserCacheKeyPrefix)
		if userID == userKey {
			continue // Skip keys without expected prefix
		}

		err := uoj.processUser(ctx, userKey, userID)
		if err != nil {
			result.errors = append(result.errors, err.Error())
		} else {
			result.offboardedCount++
		}
	}

	return result
}

// processUser handles the complete processing workflow for a single user.
//
// This method:
//  1. Retrieves user data from cache
//  2. Checks user status in LDAP
//  3. Initiates offboarding if user is inactive
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userKey: The Redis key for this user
//   - userID: The extracted user identifier
//
// Returns:
//   - error: Any error encountered during user processing, nil if successful
func (uoj *UserOffboardingJob) processUser(ctx context.Context, userKey, userID string) error {
	logger := log.FromContext(ctx)

	userData, err := uoj.getUserFromCache(ctx, userKey)
	if err != nil {
		logger.Error(err, "Failed to get user data from cache", "userKey", userKey)
		return fmt.Errorf("failed to get user %s from cache: %v", userID, err)
	}

	isActive, err := uoj.isUserActiveInLDAP(ctx, userID)
	if err != nil {
		logger.Error(err, "Failed to check LDAP status for user", "userID", userID)
		return fmt.Errorf("failed to check LDAP for user %s: %v", userID, err)
	}

	if !isActive {
		return uoj.offboardUser(ctx, userKey, userID, userData)
	}

	return nil
}

// offboardUser performs the complete offboarding process for an inactive user.
//
// This method:
//  1. Removes user from all configured backends
//  2. Deletes user data from cache
//  3. Logs the successful offboarding
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userKey: The Redis key for this user
//   - userID: The user identifier
//   - userData: The user data retrieved from cache
//
// Returns:
//   - error: Any error encountered during offboarding, nil if successful
func (uoj *UserOffboardingJob) offboardUser(ctx context.Context, userKey, userID string, userData *structs.User) error {
	logger := log.FromContext(ctx)
	logger.Info("User is inactive in LDAP, starting offboarding", "userID", userID, "email", userData.Email)

	err := uoj.offboardUserFromAllBackends(ctx, userData)
	if err != nil {
		logger.Error(err, "Failed to offboard user from backends", "userID", userID)
		return fmt.Errorf("failed to offboard user %s from backends: %v", userID, err)
	}

	// Lock cache before deletion operations to prevent concurrent modifications
	uoj.cacheMutex.Lock()
	defer uoj.cacheMutex.Unlock()

	logger.Info("Acquired cache lock for user deletion operations", "userID", userID)

	err = uoj.cacheClient.Delete(ctx, userKey)
	if err != nil {
		logger.Error(err, "Failed to remove user from cache", "userKey", userKey)
		return fmt.Errorf("failed to remove user %s from cache: %v", userID, err)
	}

	// Remove user from the user_list cache
	err = uoj.removeUserFromUserList(ctx, userID)
	if err != nil {
		logger.Error(err, "Failed to remove user from user list cache", "userID", userID)
		// Don't fail the operation, just log the error since the user is already offboarded
	}

	logger.Info("Successfully offboarded user", "userID", userID, "email", userData.Email)
	return nil
}

// getUserKeysFromCache retrieves all user keys from the cache that match the user key prefix.
//
// This method uses the cache's ScanKeys functionality to find all keys matching the
// pattern "user:*" in both Redis and in-memory cache implementations.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//
// Returns:
//   - []string: Slice of user keys found in cache matching "user:*" pattern
//   - error: Any error encountered during key retrieval
func (uoj *UserOffboardingJob) getUserKeysFromCache(ctx context.Context) ([]string, error) {
	logger := log.FromContext(ctx)
	logger.Info("Scanning cache for user keys")

	// Lock cache for read operation
	uoj.cacheMutex.RLock()
	defer uoj.cacheMutex.RUnlock()

	keys, err := uoj.cacheClient.Get(ctx, "user_list")
	if err != nil {
		return nil, fmt.Errorf("failed to scan cache for user keys: %w", err)
	}

	var userKeys []string
	keysStr, ok := keys.(string)
	if !ok {
		return nil, fmt.Errorf("user keys are not a string")
	}
	if err := json.Unmarshal([]byte(keysStr), &userKeys); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user keys: %w", err)
	}

	return userKeys, nil
}

// getUserFromCache retrieves and deserializes user data from the cache.
//
// This method fetches the JSON representation of user data from cache
// and unmarshals it into a User struct for processing.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userKey: The Redis key identifying the user data
//
// Returns:
//   - *structs.User: The deserialized user data
//   - error: Any error encountered during retrieval or unmarshaling
func (uoj *UserOffboardingJob) getUserFromCache(ctx context.Context, userKey string) (*structs.User, error) {
	// Lock cache for read operation
	uoj.cacheMutex.RLock()
	defer uoj.cacheMutex.RUnlock()

	userData, err := uoj.cacheClient.Get(ctx, userKey)
	if err != nil {
		return nil, err
	}

	userJSON, ok := userData.(string)
	if !ok {
		return nil, fmt.Errorf("user data is not a string")
	}

	var user structs.User
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// isUserActiveInLDAP verifies whether a user exists and is active in the LDAP directory.
//
// This method queries the LDAP directory for the specified user ID. If the user
// is found, they are considered active. If the user is not found (ErrNoUserFound),
// they are considered inactive and should be offboarded.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userID: The user identifier to check in LDAP
//
// Returns:
//   - bool: true if user is active in LDAP, false if inactive
//   - error: Any LDAP query error (excluding ErrNoUserFound which indicates inactivity)
func (uoj *UserOffboardingJob) isUserActiveInLDAP(ctx context.Context, userID string) (bool, error) {
	_, err := uoj.ldapClient.GetUserLDAPData(ctx, userID)
	if err != nil {
		if err == ldap.ErrNoUserFound {
			// User not found in LDAP means they're inactive
			return false, nil
		}
		// Other errors should be returned as is
		return false, err
	}

	// User found in LDAP means they're active
	return true, nil
}

// offboardUserFromAllBackends removes the specified user from selected backend systems.
//
// This method iterates through enabled backend clients and offboards users from
// all backends except GitLab and Rover, which are explicitly skipped to preserve
// access for those systems during user offboarding.
//
// Skipped backends (access preserved):
//   - GitLab: User access remains intact
//   - Rover: User access remains intact
//
// All other backend types (Fivetran, Snowflake, etc.) will have user access removed.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - user: The user data containing ID and other details for removal
//
// Returns:
//   - error: Combined error message if any backends failed, nil if all succeeded
func (uoj *UserOffboardingJob) offboardUserFromAllBackends(ctx context.Context, user *structs.User) error {
	var errors []string
	logger := log.FromContext(ctx)

	// Define which backend types should be skipped
	skippedBackendTypes := map[string]bool{
		"gitlab": true,
		"rover":  true,
	}

	for backendKey, client := range uoj.backendClients {
		// Extract backend type from the key format "{name}_{type}"
		parts := strings.Split(backendKey, "_")
		if len(parts) < 2 {
			logger.Info("Skipping backend with invalid key format", "backend", backendKey)
			continue
		}
		backendType := strings.ToLower(parts[len(parts)-1])

		// Skip backends that are explicitly excluded
		if skippedBackendTypes[backendType] {
			logger.Info("Skipping user offboarding for excluded backend type",
				"userID", user.ID, "backend", backendKey, "type", backendType)
			continue
		}

		// Proceed with offboarding for all other backends
		logger.Info("Starting user offboarding from backend",
			"userID", user.ID, "backend", backendKey, "type", backendType)

		err := client.DeleteUser(ctx, user.ID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("backend %s: %v", backendKey, err))
			logger.Error(err, "Failed to remove user from backend",
				"userID", user.ID, "backend", backendKey, "type", backendType)
			continue
		}

		logger.Info("Successfully removed user from backend",
			"userID", user.ID, "backend", backendKey, "type", backendType)
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove user from some backends: %v", errors)
	}

	return nil
}

// removeUserFromUserList removes the specified user from the user_list cache.
//
// This method retrieves the current user list from cache, removes the specified user,
// and updates the cache with the modified list. This ensures that offboarded users
// are not processed again in subsequent offboarding job runs.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userID: The ID of the user to remove from the list
//
// Returns:
//   - error: Any error encountered during the removal operation
func (uoj *UserOffboardingJob) removeUserFromUserList(ctx context.Context, userID string) error {
	logger := log.FromContext(ctx)
	logger.Info("Removing user from user list cache", "userID", userID)

	// Note: This method assumes the caller has already acquired the necessary mutex lock
	// Get current user list
	userListCache, err := uoj.cacheClient.Get(ctx, "user_list")
	if err != nil {
		return fmt.Errorf("failed to get user list from cache: %w", err)
	}

	var userList []string
	userListStr, ok := userListCache.(string)
	if !ok {
		return fmt.Errorf("user list is not a string")
	}

	if err := json.Unmarshal([]byte(userListStr), &userList); err != nil {
		return fmt.Errorf("failed to unmarshal user list: %w", err)
	}

	// Remove the user from the list
	updatedUserList := make([]string, 0, len(userList))
	for _, user := range userList {
		if user != userID {
			updatedUserList = append(updatedUserList, user)
		}
	}

	// Update the cache with the modified list
	updatedUserListJSON, err := json.Marshal(updatedUserList)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user list: %w", err)
	}

	err = uoj.cacheClient.Set(ctx, "user_list", string(updatedUserListJSON), cache.NoExpiration)
	if err != nil {
		return fmt.Errorf("failed to update user list in cache: %w", err)
	}

	logger.Info("Successfully removed user from user list cache",
		"userID", userID,
		"previousCount", len(userList),
		"newCount", len(updatedUserList))

	return nil
}
