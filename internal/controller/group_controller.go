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

package controller

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/internal/controller/controllerutils"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/gitlab"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/snowflake"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	groupFinalizer = "operator.dataverse.redhat.com/finalizer"

	// requeueAfter is the duration after which the group controller will requeue the group for reconciliation
	// this takes care of updating users in ldap query based groups
	requeueAfter = 8 * time.Hour
)

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	AppConfig       *config.AppConfig
	Store           *store.Store
	log             *logrus.Entry
	backendLogger   *logrus.Entry
	LdapConn        ldap.LDAPClient
	allLdapUserData map[string]*structs.LDAPUser

	// CacheMutex prevents concurrent access to the cache during group reconciliation.
	// This shared mutex ensures that the group controller and user offboarding job don't interfere
	// with each other when reading or modifying user/team data in Redis.
	// This mutex is shared across components and passed from main.go.
	CacheMutex *sync.RWMutex
}

//nolint:lll
// +kubebuilder:rbac:groups=operator.dataverse.redhat.com,namespace=usernaut,resources=groups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.dataverse.redhat.com,namespace=usernaut,resources=groups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.dataverse.redhat.com,namespace=usernaut,resources=groups/finalizers,verbs=update

func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = logger.WithRequestId(ctx, controller.ReconcileIDFromContext(ctx))
	r.log = logger.Logger(ctx).WithFields(logrus.Fields{
		"request": req.NamespacedName.String(),
	})

	groupCR := &usernautdevv1alpha1.Group{}

	if err := r.Get(ctx, req.NamespacedName, groupCR); err != nil {
		r.log.WithError(err).Error("Unable to fetch Group CR")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if groupCR.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, r.handleDeletion(ctx, groupCR)
	}

	// Object is not being deleted, add finalizer if missing
	if !controllerutil.ContainsFinalizer(groupCR, groupFinalizer) {
		controllerutil.AddFinalizer(groupCR, groupFinalizer)
		if err := r.Update(ctx, groupCR); err != nil {
			return ctrl.Result{}, err
		}
	}

	// set owner reference to the group CR
	if err := r.setOwnerReference(ctx, groupCR); err != nil {
		r.log.WithError(err).Error("error setting owner reference")
		return ctrl.Result{}, err
	}

	// set the group status as waiting
	groupCR.SetWaiting()
	if err := r.Status().Update(ctx, groupCR); err != nil {
		r.log.WithError(err).Error("error updating the status")
		return ctrl.Result{}, err
	}

	r.log = logger.Logger(ctx).WithFields(logrus.Fields{
		"request":        req.NamespacedName.String(),
		"group":          groupCR.Spec.GroupName,
		"has_ldap_query": groupCR.Spec.Members.LDAPQuery != nil,
		"members":        len(groupCR.Spec.Members.Users),
		"groups":         groupCR.Spec.Members.Groups,
	})

	var err error
	queryMembers := []string{}
	if groupCR.Spec.Members.LDAPQuery != nil {
		includeIndirectReports := groupCR.Spec.Members.LDAPQuery.Options != nil && groupCR.Spec.Members.LDAPQuery.Options.IncludeIndirectReports
		includeManager := groupCR.Spec.Members.LDAPQuery.Options != nil && groupCR.Spec.Members.LDAPQuery.Options.IncludeManager
		queryMembers, err = r.fetchQueryMembers(ctx, groupCR.Spec.Members.LDAPQuery, includeIndirectReports, nil)
		if err != nil {
			r.log.WithError(err).Error("error fetching query members")
			return ctrl.Result{}, err
		}
		if includeManager {
			queryMembers = append(queryMembers, extractManagerUIDsFromQuery(groupCR.Spec.Members.LDAPQuery)...)
		}
		r.log.WithField("query_members_count", len(queryMembers)).Info("query members fetched successfully")
	}

	visitedGroups := make(map[string]struct{})
	allDeclaredMembers, err := r.fetchUniqueGroupMembers(ctx, req.Name, groupCR.Namespace, visitedGroups)
	if err != nil {
		r.log.WithError(err).Error("error fetching unique group members")
		return ctrl.Result{}, err
	}

	uniqueMembers := r.deduplicateMembers(append(allDeclaredMembers, queryMembers...))

	r.log.WithField("unique_members", len(uniqueMembers)).Info("unique members to be reconciled")
	groupCR.Status.ReconciledUsers = uniqueMembers

	r.log.Info("fetching LDAP data for the users in the group")

	// Lock cache for all read/write operations during reconciliation
	// This prevents race conditions when multiple Group CRs reference the same users/teams
	// and their reconciliations run concurrently
	r.CacheMutex.Lock()
	defer r.CacheMutex.Unlock()

	r.log.Info("Acquired cache lock for entire reconciliation (LDAP + backends)")

	// Step 1: Fetch LDAP data (does NOT update cache indexes)
	ldapResult := r.fetchLDAPData(ctx, uniqueMembers)

	// Step 2: Process all backends (cache operations protected by lock)
	backendErrors := r.processAllBackends(ctx, groupCR, uniqueMembers)

	// Step 3: Only update cache indexes if ALL backends succeeded (all-or-nothing)
	hasErrors := false
	for _, m := range backendErrors {
		if len(m) > 0 {
			hasErrors = true
			break
		}
	}

	if !hasErrors {
		r.log.Info("All backends succeeded, updating cache indexes")
		if err := r.updateCacheIndexes(ctx, groupCR.Spec.GroupName, ldapResult); err != nil {
			r.log.WithError(err).Error("error updating cache indexes")
			// Continue to update status - cache index errors are logged but not fatal
		}
	} else {
		r.log.Warn("Backend errors detected, skipping cache index updates (all-or-nothing)")
	}

	// Step 4: Remove force reconcile label if present
	if removeErr := controllerutils.RemoveForceReconcileLabel(ctx, r.Client, groupCR); removeErr != nil {
		r.log.WithError(removeErr).Error("Failed to remove force reconcile label")
		return ctrl.Result{}, removeErr
	}

	// Step 5: Update status and handle errors
	if err := r.updateStatusAndHandleErrors(ctx, groupCR, backendErrors); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// LDAPFetchResult contains the results of LDAP data fetching
type LDAPFetchResult struct {
	CurrentMembers []string // emails of users with valid LDAP data
	ActiveUserList []string // UIDs of active users
}

// fetchQueryMembers runs the LDAP query and, when the query has a manager filter and
// includeIndirectReports is true, recursively expands each member's reports (people who
// report to them) and returns the combined set. visited tracks UIDs already expanded to
// avoid cycles; pass nil for the top-level call (a new map is allocated).
func (r *GroupReconciler) fetchQueryMembers(ctx context.Context, query *usernautdevv1alpha1.LDAPQuery, includeIndirectReports bool, visited map[string]struct{}) ([]string, error) {
	log := logger.Logger(ctx).WithField("fetching query members", query)

	if visited == nil {
		visited = make(map[string]struct{})
	}

	log.WithField("ldap_query", query).Info("building query string from YAML")

	queryString, err := r.LdapConn.BuildLDAPQueryFromSpec(ctx, query)
	if err != nil {
		log.WithError(err).Error("failed to build ldap query from spec")
		return nil, err
	}

	log.WithField("query_string", queryString).Info("query string built successfully")
	var queryMembers []string
	// Retry LDAP query up to 3 times for transient failures.
	for attempt := 1; attempt <= 3; attempt++ {
		queryMembers, err = r.LdapConn.GetQueryMembers(ctx, queryString)
		if err == nil {
			break
		}
		log.WithError(err).WithField("attempt", attempt).Warn("error fetching users from LDAP using the query")
		if attempt < 3 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
	}
	if err != nil {
		log.WithError(err).Error("failed to fetch users from LDAP using the query after retries")
		return nil, err
	}

	if len(queryMembers) == 0 {
		log.Info("no users found in LDAP using the query")
		return nil, nil
	}

	log.WithField("query_members_count", len(queryMembers)).Info("query members fetched successfully")

	hasManagerFilter := false
	for _, filter := range query.Filters {
		if strings.EqualFold(strings.TrimSpace(filter.Key), "manager") {
			hasManagerFilter = true
			break
		}
	}

	// Manager filter present but indirect reports disabled: return only direct reports of the manager in the query (no recursion).
	if hasManagerFilter && !includeIndirectReports {
		return r.deduplicateMembers(queryMembers), nil
	}

	if hasManagerFilter && includeIndirectReports {
		log.Info("has manager filter, fetching indirect reports")
		nestedQuery := usernautdevv1alpha1.LDAPQuery{
			Operator: query.Operator,
		}

		queue := make([]string, 0, len(queryMembers))
		queue = append(queue, queryMembers...)

		// Using DFS search based on a queue.
		// We can use this queue later for parallelization using goroutine and semaphore.
		for len(queue) > 0 {
			// Check if the context is done
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			// Use DFS
			member := queue[len(queue)-1]
			queue = queue[:len(queue)-1]

			if _, seen := visited[member]; seen {
				log.WithField("member", member).Debug("skipping already-expanded member to avoid cycle")
				continue
			}
			visited[member] = struct{}{}

			nestedQuery.Filters = make([]usernautdevv1alpha1.LDAPFilter, 0, len(query.Filters))
			for _, filter := range query.Filters {
				value := filter.Value
				if strings.EqualFold(strings.TrimSpace(filter.Key), "manager") {
					value = member
				}
				nestedQuery.Filters = append(nestedQuery.Filters, usernautdevv1alpha1.LDAPFilter{
					Key:      filter.Key,
					Criteria: filter.Criteria,
					Value:    value,
				})
			}
			nestedQueryMembers, err := r.fetchQueryMembers(ctx, &nestedQuery, includeIndirectReports, visited)
			if err != nil {
				log.WithError(err).WithField("manager", member).Error("error fetching indirect reports")
				continue
			}
			if len(nestedQueryMembers) > 0 {
				log.WithField("manager", member).WithField("reports", nestedQueryMembers).Info("reports found")
				queue = append(queue, nestedQueryMembers...)
				queryMembers = append(queryMembers, nestedQueryMembers...)
			}
		}
	}

	return r.deduplicateMembers(queryMembers), nil
}

func extractManagerUIDsFromQuery(query *usernautdevv1alpha1.LDAPQuery) []string {
	if query == nil {
		return nil
	}

	managerUIDs := make([]string, 0)
	seen := make(map[string]struct{})
	for _, filter := range query.Filters {
		if !strings.EqualFold(strings.TrimSpace(filter.Key), "manager") {
			continue
		}

		uid := filter.Value
		if uid == "" {
			continue
		}
		if _, ok := seen[uid]; ok {
			continue
		}
		seen[uid] = struct{}{}
		managerUIDs = append(managerUIDs, uid)
	}

	return managerUIDs
}

// fetchLDAPData fetches LDAP data for all unique members and populates allLdapUserData
// This function does NOT update any cache indexes - it only fetches data
// NOTE: This function assumes CacheMutex is already held by the caller
func (r *GroupReconciler) fetchLDAPData(
	ctx context.Context,
	uniqueMembers []string,
) *LDAPFetchResult {
	// Initialize LDAP user data map
	r.allLdapUserData = make(map[string]*structs.LDAPUser, len(uniqueMembers))

	// Use a map to track unique UIDs to avoid duplicates
	uniqueUIDs := make(map[string]bool)

	// Track current valid members (users with valid LDAP data)
	currentMembers := make([]string, 0, len(uniqueMembers))

	// Process each unique member - fetch LDAP data only
	for _, user := range uniqueMembers {
		ldapUserData, err := r.LdapConn.GetUserLDAPData(ctx, user)
		if err != nil {
			r.log.WithError(err).Error("error fetching user data from LDAP")
			delete(uniqueUIDs, user)
			continue
		}

		ldapUser := &structs.LDAPUser{}
		err = utils.MapToStruct(ldapUserData, ldapUser)
		if err != nil {
			r.log.WithError(err).Error("error converting LDAP user data to struct")
			continue
		}

		r.allLdapUserData[user] = ldapUser

		// Only add UID if it's not already in the list
		if !uniqueUIDs[ldapUser.GetUID()] {
			uniqueUIDs[ldapUser.GetUID()] = true
		}

		// Track this user as a current member
		email := ldapUser.GetEmail()
		currentMembers = append(currentMembers, email)
	}

	// Build list of users who are active in LDAP
	activeUserList := make([]string, 0, len(uniqueUIDs))
	for uid, isActive := range uniqueUIDs {
		if isActive {
			activeUserList = append(activeUserList, uid)
		}
	}

	return &LDAPFetchResult{
		CurrentMembers: currentMembers,
		ActiveUserList: activeUserList,
	}
}

// updateCacheIndexes updates all cache indexes after successful backend reconciliation
// This includes: user:groups reverse index, group members, and user list
// NOTE: This function assumes CacheMutex is already held by the caller
// Returns an error if critical cache updates fail
func (r *GroupReconciler) updateCacheIndexes(
	ctx context.Context,
	groupName string,
	ldapResult *LDAPFetchResult,
) error {
	var errors []error

	// Get previous members of this group (for removal detection)
	previousMembers, err := r.Store.Group.GetMembers(ctx, groupName)
	if err != nil {
		r.log.WithError(err).Warn("error fetching previous group members, assuming empty")
		previousMembers = []string{}
	}
	previousMembersSet := make(map[string]struct{}, len(previousMembers))
	for _, email := range previousMembers {
		previousMembersSet[email] = struct{}{}
	}

	// Build current members set for comparison
	currentMembersSet := make(map[string]struct{}, len(ldapResult.CurrentMembers))
	for _, email := range ldapResult.CurrentMembers {
		currentMembersSet[email] = struct{}{}
	}

	// Update user:groups reverse index - add this group to each current member's group list
	for _, email := range ldapResult.CurrentMembers {
		if err := r.Store.UserGroups.AddGroup(ctx, email, groupName); err != nil {
			r.log.WithError(err).WithField("user", email).Error("error updating user groups index")
			errors = append(errors, fmt.Errorf("failed to add group %s to user %s: %w", groupName, email, err))
		}
	}

	// Find users who were removed from the group (previous - current)
	for email := range previousMembersSet {
		if _, stillMember := currentMembersSet[email]; !stillMember {
			// User was removed from the group - update their user:groups index
			r.log.WithField("user", email).WithField("group", groupName).Info("removing group from user's group list")
			if err := r.Store.UserGroups.RemoveGroup(ctx, email, groupName); err != nil {
				r.log.WithError(err).WithField("user", email).Error("error removing group from user's groups index")
				errors = append(errors, fmt.Errorf("failed to remove group %s from user %s: %w", groupName, email, err))
			}
		}
	}

	// Update group members in consolidated store - this is critical
	if err := r.Store.Group.SetMembers(ctx, groupName, ldapResult.CurrentMembers); err != nil {
		r.log.WithError(err).Error("error updating group members")
		return fmt.Errorf("failed to update group members for %s: %w", groupName, err)
	}

	// Return combined errors if any user group index updates failed
	if len(errors) > 0 {
		return fmt.Errorf("cache index update completed with %d errors: %v", len(errors), errors)
	}

	return nil
}

// processAllBackends handles processing of all backends in the group CR
func (r *GroupReconciler) processAllBackends(
	ctx context.Context,
	groupCR *usernautdevv1alpha1.Group,
	uniqueMembers []string,
) map[string]map[string]string {
	backendErrors := make(map[string]map[string]string, 0)

	// Create a map of valid backends for validation
	validBackends := make(map[string]bool)
	for _, backend := range groupCR.Spec.Backends {
		validBackends[backend.Name+"_"+backend.Type] = true
	}

	// Group Params by backend name for direct lookup.
	groupParamsByBackend := make(map[string]structs.TeamParams)
	for _, param := range groupCR.Spec.GroupParams {
		backendKey := param.Name + "_" + param.Backend
		if !validBackends[backendKey] {
			if _, ok := backendErrors[param.Backend]; !ok {
				backendErrors[param.Backend] = make(map[string]string)
			}
			backendErrors[param.Backend][param.Name] = fmt.Errorf(
				"group param refers to non-existent backend: %s/%s",
				param.Backend, param.Name).Error()
			continue
		}
		if param.Property == "" {
			if _, ok := backendErrors[param.Backend]; !ok {
				backendErrors[param.Backend] = make(map[string]string)
			}
			backendErrors[param.Backend][param.Name] = fmt.Errorf(
				"group param property is empty for backend: %s/%s",
				param.Backend, param.Name).Error()
			continue
		} else {
			groupParamsByBackend[backendKey] = structs.TeamParams{
				Property: param.Property,
				Value:    param.Value,
			}
		}
	}

	for _, backend := range groupCR.Spec.Backends {
		r.backendLogger = r.log.WithFields(logrus.Fields{
			"backend":      backend.Name,
			"backend_type": backend.Type,
		})
		backendKey := backend.Name + "_" + backend.Type
		backendGroupParams := groupParamsByBackend[backendKey]
		if err := r.processSingleBackend(ctx, groupCR, backend, uniqueMembers, backendGroupParams); err != nil {
			r.backendLogger.WithError(err).Error("error processing backend")
			if _, ok := backendErrors[backend.Type]; !ok {
				backendErrors[backend.Type] = make(map[string]string)
			}
			backendErrors[backend.Type][backend.Name] = err.Error()
		}
	}

	return backendErrors
}

// processSingleBackend handles processing of a single backend
func (r *GroupReconciler) processSingleBackend(ctx context.Context,
	groupCR *usernautdevv1alpha1.Group,
	backend usernautdevv1alpha1.Backend,
	uniqueMembers []string,
	backendGroupParams structs.TeamParams,
) error {
	// Create backend client
	backendClient, err := clients.New(backend.Name, backend.Type, r.AppConfig.BackendMap)
	if err != nil {
		r.backendLogger.WithError(err).Error("error creating backend client")
		return err
	}
	r.backendLogger.Debug("created backend client successfully")

	isLdapSync, err := r.setupLdapSync(
		backend.Type, backend.Name, backendClient, groupCR.Spec.GroupName, groupCR.Spec.Backends,
	)
	if err != nil {
		r.backendLogger.Errorf("failed to setup ldap sync for %s: %v", backend.Type, err)
		return err
	}
	if !isLdapSync {
		r.backendLogger.Infof("ldap sync is not setup for %s backend", backend.Type)
	}

	// Fetch or create team
	backendParams := &structs.BackendParams{
		Name: backend.Name,
		Type: backend.Type,
	}
	teamID, err := r.fetchOrCreateTeam(ctx, groupCR.Spec.GroupName, backendClient, backendParams)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching or creating team")
		return err
	}
	r.backendLogger.WithField("team_id", teamID).Info("fetched or created team successfully")

	// Independent reconciliation of Group Params for each backend
	if backendGroupParams.Property != "" {
		err = backendClient.ReconcileGroupParams(ctx, teamID, backendGroupParams)
		if err != nil {
			r.backendLogger.WithError(err).Error("error reconciling group params")
			return err
		}
		r.backendLogger.Info("successfully reconciled group params")
	}

	// Create users in backend and cache
	if err := r.createUsersInBackendAndCache(ctx, uniqueMembers, backend.Name, backend.Type, backendClient); err != nil {
		r.backendLogger.WithError(err).Error("error creating users in backend and cache")
		return err
	}
	r.backendLogger.Info("created users in backend and cache successfully")

	// Fetch existing team members
	members, err := backendClient.FetchTeamMembersByTeamID(ctx, teamID)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching team members")
		return err
	}
	r.backendLogger.WithField("team_members_count", len(members)).Info("fetched team members successfully")

	// Process users (determine who to add/remove)
	usersToAdd, usersToRemove, err := r.processUsers(ctx, uniqueMembers, members, backend.Name, backend.Type)
	if err != nil {
		r.backendLogger.WithError(err).Error("error processing users")
		return err
	}

	// Add users to team if needed
	if !isLdapSync {
		if len(usersToAdd) > 0 {
			r.backendLogger.WithField("user_count", len(usersToAdd)).Info("Adding users to the team")
			if err := backendClient.AddUserToTeam(ctx, teamID, usersToAdd); err != nil {
				r.backendLogger.WithError(err).Error("error while adding users to the team")
				return err
			}
			r.backendLogger.WithField("users_to_add", usersToAdd).Info("added users to team successfully")
		}

		// Remove users from team if needed
		if len(usersToRemove) > 0 {
			r.backendLogger.WithField("user_count", len(usersToRemove)).Info("removing users from a team")
			if err := backendClient.RemoveUserFromTeam(ctx, teamID, usersToRemove); err != nil {
				r.backendLogger.WithError(err).Error("error while removing users from the team")
				return err
			}
			r.backendLogger.WithField("users_to_remove", usersToRemove).Info("removed users from team successfully")
		}
	}

	r.backendLogger.Info("successfully processed backend")

	return nil
}

// updateStatusAndHandleErrors updates the CR status and handles any backend errors
func (r *GroupReconciler) updateStatusAndHandleErrors(ctx context.Context,
	groupCR *usernautdevv1alpha1.Group,
	backendErrors map[string]map[string]string) error {
	backendStatus := make([]usernautdevv1alpha1.BackendStatus, 0, len(groupCR.Spec.Backends))

	// Build status for each backend
	for _, backend := range groupCR.Spec.Backends {
		status := usernautdevv1alpha1.BackendStatus{
			Name: backend.Name,
			Type: backend.Type,
		}
		if typeMap, ok := backendErrors[backend.Type]; ok {
			if msg, found := typeMap[backend.Name]; found {
				status.Status = false
				status.Message = msg
			} else {
				status.Status = true
				status.Message = "Successful"
			}
		} else {
			status.Status = true
			status.Message = "Successful"
		}
		backendStatus = append(backendStatus, status)
	}

	// Update CR status
	groupCR.Status.BackendsStatus = backendStatus
	groupCR.UpdateStatus(false)
	hasErrors := false
	for _, m := range backendErrors {
		if len(m) > 0 {
			hasErrors = true
			break
		}
	}
	if hasErrors {
		groupCR.UpdateStatus(true)
	}
	if updateStatusErr := r.Status().Update(ctx, groupCR); updateStatusErr != nil {
		r.log.WithError(updateStatusErr).Error("error while updating final status")
		return updateStatusErr
	}

	// Return error if any backends failed
	if hasErrors {
		return errors.New("failed to reconcile all backends")
	}

	return nil
}

// handleDeletion processes the deletion of a Group CR and its finalizer
func (r *GroupReconciler) handleDeletion(ctx context.Context, groupCR *usernautdevv1alpha1.Group) error {
	if controllerutil.ContainsFinalizer(groupCR, groupFinalizer) {
		// Lock cache for deletion operations
		// Multiple Group CRs might reference the same team and delete concurrently
		r.CacheMutex.Lock()
		defer r.CacheMutex.Unlock()

		// Clean up user:groups reverse index for all members of this group
		r.cleanupUserGroupsIndex(ctx, groupCR.Spec.GroupName)

		if err := r.deleteBackendsTeam(ctx, groupCR); err != nil {
			return err
		}

		controllerutil.RemoveFinalizer(groupCR, groupFinalizer)
		if err := r.Update(ctx, groupCR); err != nil {
			r.log.WithError(err).Error("error while updating group CR")
			return err
		}
	}
	return nil
}

// cleanupUserGroupsIndex removes the group from all members' user:groups index
// NOTE: Caller must hold CacheMutex lock
// NOTE: This does NOT delete the group entry - that happens in deleteBackendsTeam
func (r *GroupReconciler) cleanupUserGroupsIndex(ctx context.Context, groupName string) {
	// Get all members of the group
	members, err := r.Store.Group.GetMembers(ctx, groupName)
	if err != nil {
		r.log.WithError(err).Warn("error fetching group members for cleanup")
		return // Nothing to clean up
	}

	// Remove the group from each member's user:groups index
	for _, email := range members {
		r.log.WithFields(logrus.Fields{
			"user":  email,
			"group": groupName,
		}).Info("removing group from user's group list during deletion")
		if err := r.Store.UserGroups.RemoveGroup(ctx, email, groupName); err != nil {
			r.log.WithError(err).WithField("user", email).Error("error removing group from user's groups index during deletion")
			// Continue processing other members
		}
	}

	r.log.WithField("group", groupName).Info("cleaned up user groups index successfully")
}

func (r *GroupReconciler) deleteBackendsTeam(ctx context.Context, groupCR *usernautdevv1alpha1.Group) error {
	r.log.Info("Finalizer: starting Backends team deletion cleanup")
	groupName := groupCR.Spec.GroupName

	for _, backend := range groupCR.Spec.Backends {
		transformedGroupName, err := utils.GetTransformedGroupName(r.AppConfig, backend.Type, groupName)
		backendLoggerInfo := r.log.WithFields(logrus.Fields{
			"group_name":            groupName,
			"transformed_team_name": transformedGroupName,
			"backend":               backend.Name,
			"backend_type":          backend.Type,
		})
		backendLoggerInfo.Info("Finalizer: Deleting team from backend")
		if err != nil {
			backendLoggerInfo.WithError(err).Error("Finalizer: Error in transforming group name")
			return err
		}

		backendClient, err := clients.New(backend.Name, backend.Type, r.AppConfig.BackendMap)
		if err != nil {
			backendLoggerInfo.WithError(err).Errorf("Finalizer: error creating client for backend %s", backend.Name)
			return err
		}

		// Get team ID from consolidated group store (using original group name)
		// NOTE: CacheMutex is already held by caller (handleDeletion)
		teamID, err := r.Store.Group.GetBackendID(ctx, groupName, backend.Name, backend.Type)
		if err != nil {
			backendLoggerInfo.WithError(err).Error("Finalizer: error fetching team details from cache")
			return err
		}

		if teamID != "" {
			backendLoggerInfo.Infof("Finalizer: Deleting team with (ID: %s) from Backend %s", teamID, backend.Type)

			if err := backendClient.DeleteTeamByID(ctx, teamID); err != nil {
				backendLoggerInfo.WithError(err).Error("Finalizer: failed to delete team from the backend")
				return err
			}
			backendLoggerInfo.Infof("Finalizer: Successfully deleted team with id '%s' from Backend %s", teamID, backend.Type)
		}

		// Delete team entry from TeamStore (used for preload lookups)
		if err := r.Store.Team.Delete(ctx, transformedGroupName); err != nil {
			backendLoggerInfo.WithError(err).Warn("Finalizer: failed to delete team from TeamStore cache")
			// Continue processing - TeamStore is secondary cache
		}
	}

	// Delete the entire group entry from cache (includes all backends and members)
	if err := r.Store.Group.Delete(ctx, groupName); err != nil {
		r.log.WithError(err).Error("Finalizer: failed to delete group from cache")
		return err
	}
	r.log.WithField("group", groupName).Info("Finalizer: Successfully deleted group from cache")

	return nil
}

func (r *GroupReconciler) processUsers(ctx context.Context,
	groupUsers []string,
	existingTeamMembers map[string]*structs.User,
	backendName, backendType string) ([]string, []string, error) {

	userIDsToSync := make([]string, 0)
	usersToAdd := make([]string, 0)
	usersToRemove := make([]string, 0)

	for _, user := range groupUsers {
		userDetails := r.allLdapUserData[user]
		if userDetails == nil {
			r.backendLogger.WithField("user", user).Warn("user not found in LDAP data, skipping processing for this user")

			// we need to check if the user is already in the existing team members
			if _, exists := existingTeamMembers[user]; exists {
				r.backendLogger.WithField("user", user).Info("user is already in existing team members, skipping user creation")
				usersToRemove = append(usersToRemove, user)
			}
			continue
		}

		// NOTE: CacheMutex is already held by caller (Reconcile)
		// Get user backends from cache
		userBackends, err := r.Store.User.GetBackends(ctx, userDetails.GetEmail())
		if err != nil {
			r.backendLogger.WithError(err).Error("error fetching user details from cache")
			return nil, nil, err
		}

		backendKey := backendName + "_" + backendType
		userID := userBackends[backendKey]
		if userID == "" {
			r.backendLogger.WithField("user", user).Warn("user ID not found in cache, will create user in backend")
			return nil, nil, errors.New("user ID not found in cache")
		}
		userIDsToSync = append(userIDsToSync, userID)
	}

	// process existing team members to find users to remove
	for userID := range existingTeamMembers {
		if !slices.Contains(userIDsToSync, userID) {
			usersToRemove = append(usersToRemove, userID)
		}
	}

	// process group users to find users to add
	// if user is not present in existing team members, then add the user to the team
	for _, userID := range userIDsToSync {
		if _, exists := existingTeamMembers[userID]; !exists {
			usersToAdd = append(usersToAdd, userID)
		}
	}

	return usersToAdd, usersToRemove, nil
}

// userCreateRequest holds the data needed to create a single user in a backend.
type userCreateRequest struct {
	userName string
	email    string
	user     *structs.User
}

// userCreateResult holds the outcome of a single backend CreateUser call.
type userCreateResult struct {
	email string
	id    string
}

func (r *GroupReconciler) createUsersInBackendAndCache(ctx context.Context,
	users []string,
	backendName, backendType string,
	backendClient clients.Client) error {

	// NOTE: CacheMutex is already held by caller (Reconcile)
	backendKey := backendName + "_" + backendType

	// Phase 1 (sequential): read cache, collect users that need creation
	var toCreate []userCreateRequest
	for _, user := range users {
		userDetails := r.allLdapUserData[user]
		if userDetails == nil {
			r.backendLogger.WithField("user", user).Warn("user not found in LDAP data, skipping user creation")
			continue
		}

		userBackends, err := r.Store.User.GetBackends(ctx, userDetails.GetEmail())
		if err != nil {
			r.backendLogger.WithField("user", user).WithError(err).Error("error fetching user details from cache")
			return err
		}

		if userID, exists := userBackends[backendKey]; exists && userID != "" {
			r.backendLogger.WithField("user", user).Debug("user already exists in cache")
			continue
		}

		toCreate = append(toCreate, userCreateRequest{
			userName: user,
			email:    userDetails.GetEmail(),
			user: &structs.User{
				Email:     userDetails.GetEmail(),
				UserName:  user,
				Role:      fivetran.AccountReviewerRole,
				FirstName: utils.StandardizeNameForBackend(userDetails.GetDisplayName()),
				LastName:  utils.StandardizeNameForBackend(userDetails.GetSN()),
			},
		})
	}

	if len(toCreate) == 0 {
		return nil
	}

	r.backendLogger.WithField("users_to_create", len(toCreate)).Info("creating users in backend")

	// Phase 2: call backend CreateUser API.
	results := make([]userCreateResult, len(toCreate))

	if sfClient, ok := backendClient.(*snowflake.SnowflakeClient); ok {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(sfClient.GetConfig().MaxConcurrency)

		for i, req := range toCreate {
			idx := i
			cr := req
			g.Go(func() error {
				newUser, err := backendClient.CreateUser(gctx, cr.user)
				if err != nil {
					return err
				}
				results[idx] = userCreateResult{email: cr.email, id: newUser.ID}
				return nil
			})
		}

		if err := g.Wait(); err != nil {
			return err
		}
	} else {
		for i, req := range toCreate {
			newUser, err := backendClient.CreateUser(ctx, req.user)
			if err != nil {
				return err
			}
			results[i] = userCreateResult{email: req.email, id: newUser.ID}
		}
	}

	// Phase 3 (sequential): write results back to cache under the existing lock
	for _, res := range results {
		if res.id == "" {
			continue
		}
		if err := r.Store.User.SetBackend(ctx, res.email, backendKey, res.id); err != nil {
			return err
		}
	}
	return nil
}

func (r *GroupReconciler) fetchOrCreateTeam(ctx context.Context,
	groupName string, backendClient clients.Client,
	backendParams *structs.BackendParams) (string, error) {

	backendName := backendParams.GetName()
	backendType := backendParams.GetType()

	// Get transformed group name for backend API calls (team name in backend system)
	transformedGroupName, err := utils.GetTransformedGroupName(r.AppConfig, backendType, groupName)
	if err != nil {
		r.backendLogger.WithError(err).Error("error transforming the group Name")
		return "", err
	}

	backendKey := backendName + "_" + backendType

	// Step 1: Check GroupStore first (using original group name)
	teamID, err := r.Store.Group.GetBackendID(ctx, groupName, backendName, backendType)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching team details from GroupStore")
		return "", err
	}

	if teamID != "" {
		r.backendLogger.WithField("teamID", teamID).Info("team details found in GroupStore")
		return teamID, nil
	}

	// Step 2: Fallback to TeamStore (using transformed name, populated during preload)
	teamBackends, err := r.Store.Team.GetBackends(ctx, transformedGroupName)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching team details from TeamStore")
		return "", err
	}

	if id, exists := teamBackends[backendKey]; exists && id != "" {
		r.backendLogger.WithField("teamID", id).Info("team details found in TeamStore, migrating to GroupStore")

		// Migrate data from TeamStore to GroupStore
		if err := r.Store.Group.SetBackend(ctx, groupName, backendName, backendType, id); err != nil {
			r.backendLogger.WithError(err).Error("error migrating team details to GroupStore")
			return "", err
		}

		r.backendLogger.Info("successfully migrated team details from TeamStore to GroupStore")
		return id, nil
	}

	// Step 3: Team not found in either store, create a new team
	r.backendLogger.Info("team details not found in cache, creating a new team")

	newTeam, err := backendClient.CreateTeam(ctx, &structs.Team{
		Name:        transformedGroupName, // Use transformed name for backend API
		Description: "team for " + groupName,
		Role:        fivetran.AccountReviewerRole,
	})
	if err != nil {
		r.backendLogger.WithError(err).Error("error creating team in backend")
		return "", err
	}

	r.backendLogger.Info("created team in backend successfully")

	// Store in GroupStore only - TeamStore is populated by preloadCache and used as read-only fallback
	if err := r.Store.Group.SetBackend(ctx, groupName, backendName, backendType, newTeam.ID); err != nil {
		r.backendLogger.WithError(err).Error("error updating team details in GroupStore")
		return "", err
	}

	r.backendLogger.Info("updated team details in GroupStore successfully")

	return newTeam.ID, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Add an index field for referenced groups
	indexField := "spec.members.groups"
	groupType := &usernautdevv1alpha1.Group{}
	indexFunc := func(obj client.Object) []string {
		group := obj.(*usernautdevv1alpha1.Group)
		return group.Spec.Members.Groups
	}
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), groupType, indexField, indexFunc); err != nil {
		return err
	}

	// Create a mapping function to find all Group CRs that reference a changed Group CR
	mapFunc := func(ctx context.Context, obj client.Object) []reconcile.Request {
		group := obj.(*usernautdevv1alpha1.Group)
		var referencingGroups usernautdevv1alpha1.GroupList

		// Find all Group CRs that reference this Group in their spec.members.groups
		if err := r.List(ctx, &referencingGroups, client.MatchingFields{
			indexField: group.Name,
		}); err != nil {
			r.log.WithError(err).Error("error listing referencing groups")
			return nil
		}

		// Create reconcile requests for each referencing Group
		var requests []reconcile.Request
		for _, referencingGroup := range referencingGroups.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      referencingGroup.Name,
					Namespace: referencingGroup.Namespace,
				},
			})
		}
		return requests
	}

	// force reconcile flag
	labelPredicate := controllerutils.ForceReconcilePredicate()

	maxConcurrentReconciles := r.AppConfig.ControllerConfig.MaxConcurrentReconciles
	if maxConcurrentReconciles <= 0 {
		maxConcurrentReconciles = 1 // default value
	}

	// Log the configured concurrency level
	logger.Logger(context.Background()).WithFields(logrus.Fields{
		"maxConcurrentReconciles": maxConcurrentReconciles,
	}).Info("Configuring MaxConcurrentReconciles for Group controller")

	return ctrl.NewControllerManagedBy(mgr).
		For(&usernautdevv1alpha1.Group{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, labelPredicate)).
		Watches(
			client.Object(&usernautdevv1alpha1.Group{}),
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: maxConcurrentReconciles,
		}).
		Complete(r)
}

func (r *GroupReconciler) fetchUniqueGroupMembers(ctx context.Context, groupName,
	namespace string, visitedOnPath map[string]struct{}) ([]string, error) {

	r.log.WithField("group", groupName).Info("fetching group members")

	// Handle cyclic dependencies for the current recursion path.
	if _, ok := visitedOnPath[groupName]; ok {
		r.log.WithField("group", groupName).Warn("cyclic group dependency detected; returning empty member list")
		return []string{}, nil
	}
	visitedOnPath[groupName] = struct{}{}
	defer delete(visitedOnPath, groupName) // Remove from path when returning.

	groupCR := &usernautdevv1alpha1.Group{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: groupName}, groupCR); err != nil {
		r.log.WithError(err).Error("error fetching the group CR")
		return nil, err
	}

	members := make([]string, 0)
	members = append(members, groupCR.Spec.Members.Users...)

	for _, subGroup := range groupCR.Spec.Members.Groups {
		subMembers, err := r.fetchUniqueGroupMembers(ctx, subGroup, namespace, visitedOnPath)
		if err != nil {
			return nil, err
		}
		members = append(members, subMembers...)
	}

	return members, nil
}

func (r *GroupReconciler) deduplicateMembers(members []string) []string {
	// Deduplicate groupMembers before setting status
	uniqueMembersMap := make(map[string]struct{})
	uniqueMembers := make([]string, 0, len(members))
	for _, member := range members {
		if _, exists := uniqueMembersMap[member]; !exists {
			uniqueMembersMap[member] = struct{}{}
			uniqueMembers = append(uniqueMembers, member)
		}
	}
	return uniqueMembers
}

func (r *GroupReconciler) setOwnerReference(ctx context.Context, groupCR *usernautdevv1alpha1.Group) error {
	// Determine the desired owner references from parent groups
	desiredOwnerRefs := make(map[types.UID]metav1.OwnerReference)
	for _, parentGroupName := range groupCR.Spec.Members.Groups {
		parentGroupCR := &usernautdevv1alpha1.Group{}
		if err := r.Client.Get(ctx,
			client.ObjectKey{Namespace: groupCR.Namespace, Name: parentGroupName}, parentGroupCR); err != nil {
			r.log.WithError(err).Error("error fetching the parent group CR")
			return err
		}
		blockOwnerDeletion := true
		desiredOwnerRefs[parentGroupCR.UID] = metav1.OwnerReference{
			APIVersion:         usernautdevv1alpha1.GroupVersion.String(),
			Kind:               "Group",
			Name:               parentGroupCR.Name,
			UID:                parentGroupCR.UID,
			BlockOwnerDeletion: &blockOwnerDeletion,
		}
	}

	// Separate existing owner references into Group and non-Group kinds
	var nonGroupOwnerRefs []metav1.OwnerReference
	existingGroupOwnerRefs := make(map[types.UID]struct{})
	for _, ref := range groupCR.OwnerReferences {
		if ref.Kind == "Group" && ref.APIVersion == usernautdevv1alpha1.GroupVersion.String() {
			existingGroupOwnerRefs[ref.UID] = struct{}{}
		} else {
			nonGroupOwnerRefs = append(nonGroupOwnerRefs, ref)
		}
	}

	// Check if an update is needed by comparing desired and existing Group owner references
	needsUpdate := false
	if len(desiredOwnerRefs) != len(existingGroupOwnerRefs) {
		needsUpdate = true
	} else {
		for uid := range desiredOwnerRefs {
			if _, ok := existingGroupOwnerRefs[uid]; !ok {
				needsUpdate = true
				break
			}
		}
	}

	if !needsUpdate {
		return nil
	}

	// Construct the new list of owner references and update the CR
	newOwnerRefs := make([]metav1.OwnerReference, 0, len(desiredOwnerRefs)+len(nonGroupOwnerRefs))
	newOwnerRefs = append(newOwnerRefs, nonGroupOwnerRefs...)
	for _, ref := range desiredOwnerRefs {
		newOwnerRefs = append(newOwnerRefs, ref)
	}

	groupCR.OwnerReferences = newOwnerRefs
	if err := r.Update(ctx, groupCR); err != nil {
		r.log.WithError(err).Error("error updating the group CR with owner reference")
		return err
	}

	return nil
}

func (r *GroupReconciler) setupLdapSync(backendType string,
	backendName string,
	backendClient clients.Client,
	groupName string,
	backends []usernautdevv1alpha1.Backend,
) (bool, error) {
	switch backendType {
	case "gitlab":
		dependsOn := r.AppConfig.BackendMap["gitlab"][backendName].DependsOn

		if dependsOn.Type == "" && dependsOn.Name == "" {
			r.backendLogger.Infof("no ldap dependant found for %s backend", dependsOn.Type)
			return false, nil
		}

		// Check if the dependent backend exists in cache (using original group name)
		err := r.ldapDependantChecks(dependsOn, groupName)
		if err != nil {
			return false, err
		}

		if !isGroupCRHasDependants(backends, dependsOn) {
			return false, fmt.Errorf("ldap dependants for %s backend doesn't exist in group CR", backendType)
		}

		gitlabClient, ok := backendClient.(*gitlab.GitlabClient)
		if !ok {
			return false, errors.New("backend client is not a GitlabClient")
		}
		gitlabClient.SetLdapSync(true, groupName)
		r.backendLogger.Infof("ldap sync setup successfully for %s", backendType)
		return true, nil
	}
	return false, nil
}

func (r *GroupReconciler) ldapDependantChecks(dependsOn config.Dependant, groupName string) error {
	dependantType, ok := r.AppConfig.BackendMap[dependsOn.Type]
	if !ok {
		return fmt.Errorf("ldap dependant type %s not found in BackendMap", dependsOn.Type)
	}
	dependantName, ok := dependantType[dependsOn.Name]
	if !ok {
		return fmt.Errorf("ldap dependant name %s not found in BackendMap[%s]", dependsOn.Name, dependsOn.Type)
	}
	if !dependantName.Enabled {
		return fmt.Errorf("%s is not enabled", dependsOn.Type)
	}

	// Check if the group exists in cache with the dependent backend configured
	// NOTE: This is called without holding CacheMutex (called from ldap sync)

	// First check GroupStore (using original group name)
	exists, err := r.Store.Group.BackendExists(context.Background(), groupName, dependsOn.Name, dependsOn.Type)
	if err == nil && exists {
		return nil
	}

	// Fallback to TeamStore (using transformed name)
	transformedGroupName, err := utils.GetTransformedGroupName(r.AppConfig, dependsOn.Type, groupName)
	if err != nil {
		r.backendLogger.WithError(err).Error("error transforming group name for ldap dependant check")
		return err
	}

	backendKey := dependsOn.Name + "_" + dependsOn.Type
	teamBackends, err := r.Store.Team.GetBackends(context.Background(), transformedGroupName)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching team from TeamStore for ldap dependant check")
		return err
	}

	if _, found := teamBackends[backendKey]; found {
		return nil
	}

	r.backendLogger.Error("dependent backend not found in cache for group, skipping ldap sync")
	return fmt.Errorf("dependent backend %s not found in cache for group %s", backendKey, groupName)
}

func isGroupCRHasDependants(backends []usernautdevv1alpha1.Backend, dependsOn config.Dependant) bool {
	for _, backend := range backends {
		if backend.Type == dependsOn.Type && backend.Name == dependsOn.Name {
			return true
		}
	}
	return false
}
