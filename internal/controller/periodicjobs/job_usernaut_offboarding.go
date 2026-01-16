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
	"fmt"
	"strings"
	"sync"
	"time"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/google/uuid"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// UserOffboardingJobName is the unique identifier for the user offboarding periodic job.
	UserOffboardingJobName = "usernaut_user_offboarding"

	// UserOffboardingJobInterval defines how often the user offboarding job runs.
	// Set to 24 hours to perform daily cleanup of inactive users.
	UserOffboardingJobInterval = 24 * time.Hour
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

	// store provides access to the store layer with prefixed keys
	store *store.Store

	// ldapClient enables verification of user status in the LDAP directory.
	ldapClient ldap.LDAPClient

	// backendClients contains all configured backend clients (Fivetran, Rover, etc.)
	// mapped by their unique identifier "{name}_{type}".
	backendClients map[string]clients.Client

	// cacheMutex prevents concurrent access to the cache during user offboarding operations.
	// This shared mutex ensures that the GroupReconciler and UserOffboardingJob don't interfere
	// with each other when reading or modifying user data in Redis.
	// This mutex is shared across components and passed from main.go.
	cacheMutex *sync.RWMutex

	logger *logrus.Entry
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
//   - sharedCacheMutex: Shared mutex to prevent race conditions with other components
//   - dataStore: Shared store instance with prefixed keys
//   - ldapClient: Shared LDAP client instance
//   - backendClients: Map of initialized backend clients
//
// Returns:
//   - *UserOffboardingJob: A configured job instance
func NewUserOffboardingJob(
	sharedCacheMutex *sync.RWMutex,
	dataStore *store.Store,
	ldapClient ldap.LDAPClient,
	backendClients map[string]clients.Client,
) *UserOffboardingJob {
	return &UserOffboardingJob{
		store:          dataStore,
		ldapClient:     ldapClient,
		backendClients: backendClients,
		cacheMutex:     sharedCacheMutex,
	}
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
	ctx = logger.WithRequestId(ctx, types.UID(uuid.New().String()))
	uoj.logger = logger.Logger(ctx).WithFields(logrus.Fields{
		"job": UserOffboardingJobName,
	})
	uoj.logger.Info("Starting user offboarding job")

	userKeys, err := uoj.getUserListFromCache(ctx)
	if err != nil {
		uoj.logger.Error(err, "Failed to get user keys from cache")
		return err
	}

	uoj.logger.WithField("count", len(userKeys)).Info("Found users in cache")

	result := uoj.processUsers(ctx, userKeys)

	uoj.logger.WithFields(logrus.Fields{
		"totalUsers":      len(userKeys),
		"offboardedUsers": result.offboardedCount,
		"errors":          len(result.errors),
	}).Info("User offboarding job completed")

	// Log summary table of offboarded users
	if len(result.offboardedUsers) > 0 {
		uoj.logOffboardedUsersSummary(result.offboardedUsers)
	}

	if len(result.errors) > 0 {
		return fmt.Errorf("user offboarding completed with %d errors: %v", len(result.errors), result.errors)
	}

	return nil
}

// processingResult holds the results of processing multiple users during a job execution.
type processingResult struct {
	// offboardedCount tracks the number of users successfully offboarded
	offboardedCount int
	// offboardedUsers contains the list of users that were successfully offboarded
	offboardedUsers []string
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
		uoj.logger.WithField("userKey", userKey).Debug("Processing user")
		offboarded, err := uoj.processUser(ctx, userKey)
		if err != nil {
			result.errors = append(result.errors, err.Error())
		} else if offboarded {
			result.offboardedCount++
			result.offboardedUsers = append(result.offboardedUsers, userKey)
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
//   - bool: true if user was offboarded, false if user is still active
//   - error: Any error encountered during user processing, nil if successful
func (uoj *UserOffboardingJob) processUser(ctx context.Context, userKey string) (bool, error) {
	isActive, err := uoj.isUserActiveInLDAP(ctx, userKey)
	if err != nil {
		uoj.logger.Error(err, "Failed to check LDAP status for user", "userKey", userKey)
		return false, fmt.Errorf("failed to check LDAP for user %s: %v", userKey, err)
	}

	if !isActive {
		uoj.logger.WithField("userKey", userKey).Info("User is inactive in LDAP, starting offboarding")
		err = uoj.offboardUser(ctx, userKey)
		if err != nil {
			uoj.logger.WithField("userKey", userKey).Error(err, "Failed to offboard user")
			return false, fmt.Errorf("failed to offboard user %s: %v", userKey, err)
		}
		uoj.logger.WithField("userKey", userKey).Info("Successfully offboarded user")
		return true, nil
	}
	return false, nil
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
func (uoj *UserOffboardingJob) offboardUser(ctx context.Context, userKey string) error {
	userData, userEmail, err := uoj.getUserDataFromCache(ctx, userKey)
	if err != nil {
		return fmt.Errorf("failed to get user data from cache: %w", err)
	}
	err = uoj.offboardUserFromAllBackends(ctx, userKey, userData)
	if err != nil {
		uoj.logger.WithField("userKey", userKey).Error(err, "Failed to offboard user from backends")
		return fmt.Errorf("failed to offboard user %s from backends: %v", userKey, err)
	}

	// Lock cache before deletion operations to prevent concurrent modifications
	uoj.cacheMutex.Lock()
	defer uoj.cacheMutex.Unlock()

	uoj.logger.WithField("userKey", userKey).Info("Acquired cache lock for user deletion operations")

	err = uoj.store.User.Delete(ctx, userEmail)
	if err != nil {
		uoj.logger.Error(err, "Failed to remove user from cache", "userKey", userKey, "userEmail", userEmail)
		return fmt.Errorf("failed to remove user %s from cache: %v", userKey, err)
	}

	uoj.logger.WithField("userKey", userKey).Info("Successfully offboarded user")
	return nil
}

// logOffboardedUsersSummary logs a structured summary of all offboarded users using logrus fields.
//
// This method creates structured log entries showing all users that were successfully
// removed during the offboarding job execution.
//
// Parameters:
//   - offboardedUsers: List of users that were successfully offboarded
func (uoj *UserOffboardingJob) logOffboardedUsersSummary(offboardedUsers []string) {
	if len(offboardedUsers) == 0 {
		return
	}

	uoj.logger.WithFields(logrus.Fields{
		"removedUsers": offboardedUsers,
		"totalCount":   len(offboardedUsers),
	}).Info("Offboarded users summary")
}

// getUserListFromCache retrieves all user keys from the cache that match the user key prefix.
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
func (uoj *UserOffboardingJob) getUserListFromCache(ctx context.Context) ([]string, error) {
	uoj.logger.Info("Scanning cache for user keys")

	// Lock cache for read operation
	uoj.cacheMutex.RLock()
	defer uoj.cacheMutex.RUnlock()

	userMap, err := uoj.store.User.GetByPattern(ctx, "*")
	if err != nil {
		return nil, fmt.Errorf("failed to get user list from cache: %w", err)
	}

	userKeys := make([]string, 0, len(userMap))
	for user := range userMap {
		if !strings.HasPrefix(user, "groups:") {
			userKeys = append(userKeys, user)
		}
	}

	return userKeys, nil
}

// getUserFromCache retrieves and deserializes user data from the cache.
//
// This method fetches the JSON representation of user data from cache
// and unmarshals it into a User struct for processing. It supports both exact key
// matching and pattern-based searching for email patterns.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userKey: The username to search for in cache (e.g., "subhatta" will match "subhatta@redhat.com")
//
// Returns:
//   - map[string]string: The backend mappings for the user (backend_name_type -> user_id)
//   - error: Any error encountered during retrieval or unmarshaling
func (uoj *UserOffboardingJob) getUserDataFromCache(
	ctx context.Context, userKey string,
) (map[string]string, string, error) {
	// Lock cache for read operation
	uoj.cacheMutex.RLock()
	defer uoj.cacheMutex.RUnlock()

	// userKey is a username (e.g., "subhatta"), search for cache keys that contain this username
	// We don't know the exact email, so we search broadly and then filter
	// Pattern: *username* (e.g., "*subhatta*" matches emails like "subhatta@example.com")
	usernamePattern := fmt.Sprintf("*%s*", userKey)

	userDataList, err := uoj.store.User.GetByPattern(ctx, usernamePattern)
	if err != nil {
		return nil, "", err
	}

	// Search through all matching cache entries to find the user's backend mappings
	for email, backendMap := range userDataList {
		// Return the first valid match
		return backendMap, email, nil
	}

	return nil, "", fmt.Errorf("No user found with username: %s", userKey)
}

// isUserActiveInLDAP verifies whether a user exists and is active in the LDAP directory.
//
// This method queries the LDAP directory for the specified user ID. If the user
// is found, they are considered active. If the user is not found (ErrNoUserFound),
// they are considered inactive and should be offboarded.
//
// Parameters:
//   - ctx: Context for cancellation and logging
//   - userEmail: The user identifier to check in LDAP
//
// Returns:
//   - bool: true if user is active in LDAP, false if inactive
//   - error: Any LDAP query error (excluding ErrNoUserFound which indicates inactivity)
func (uoj *UserOffboardingJob) isUserActiveInLDAP(ctx context.Context, userEmail string) (bool, error) {
	userData, err := uoj.ldapClient.GetUserLDAPDataByEmail(ctx, userEmail)
	if err != nil {
		if err == ldap.ErrNoUserFound {
			// User not found in LDAP means they're inactive
			return false, nil
		}
		// Handle LDAP "No Such Object" error using proper typed error checking
		if ldapErr, ok := err.(*goldap.Error); ok && ldapErr.ResultCode == goldap.LDAPResultNoSuchObject {
			return false, nil
		}
		// Other errors should be returned as is
		return false, err
	}

	// Check if userData is empty - treat as inactive user
	if len(userData) == 0 {
		uoj.logger.WithField("userEmail", userEmail).Info("User data is empty, treating as inactive")
		return false, nil
	}

	// User found in LDAP with valid data means they're active
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
func (uoj *UserOffboardingJob) offboardUserFromAllBackends(
	ctx context.Context, userKey string, userData map[string]string,
) error {
	var errors []string

	// Define which backend types should be skipped
	skippedBackendTypes := map[string]bool{
		"gitlab": true,
		"rover":  true,
	}

	for backendKey, client := range uoj.backendClients {
		// Extract backend type from the key format "{name}_{type}"
		parts := strings.Split(backendKey, "_")
		if len(parts) < 2 {
			uoj.logger.WithField("backend", backendKey).Info("Skipping backend with invalid key format")
			continue
		}
		backendType := strings.ToLower(parts[len(parts)-1])

		// Skip backends that are explicitly excluded
		if skippedBackendTypes[backendType] {
			uoj.logger.WithFields(logrus.Fields{
				"userKey": userKey,
				"backend": backendKey,
				"type":    backendType,
			}).Info("Skipping user offboarding for excluded backend type")
			continue
		}

		// Get the user ID for this specific backend from the userData map
		userIDStr, exists := userData[backendKey]
		if !exists {
			uoj.logger.WithFields(logrus.Fields{
				"userKey": userKey,
				"backend": backendKey,
				"type":    backendType,
			}).Info("User not found in backend, skipping")
			continue
		}

		// Proceed with offboarding for all other backends
		uoj.logger.WithFields(logrus.Fields{
			"userKey":       userKey,
			"backendUserID": userIDStr,
			"backend":       backendKey,
			"type":          backendType,
		}).Info("Starting user offboarding from backend")

		err := client.DeleteUser(ctx, userIDStr)
		if err != nil {
			errors = append(errors, fmt.Sprintf("backend %s: %v", backendKey, err))
			uoj.logger.WithFields(logrus.Fields{
				"userKey":       userKey,
				"backendUserID": userIDStr,
				"backend":       backendKey,
				"type":          backendType,
			}).Error(err, "Failed to remove user from backend")
			continue
		}

		uoj.logger.WithFields(logrus.Fields{
			"userKey":       userKey,
			"backendUserID": userIDStr,
			"backend":       backendKey,
			"type":          backendType,
		}).Info("Successfully removed user from backend")
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove user from some backends: %v", errors)
	}

	return nil
}
