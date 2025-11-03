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
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

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

	"github.com/go-redis/redis"
	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/gitlab"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	groupFinalizer = "operator.dataverse.redhat.com/finalizer"
)

// GroupReconciler reconciles a Group object
type GroupReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	AppConfig       *config.AppConfig
	Cache           cache.Cache
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
		return r.handleDeletion(ctx, groupCR)
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
		"request": req.NamespacedName.String(),
		"group":   groupCR.Spec.GroupName,
		"members": len(groupCR.Spec.Members.Users),
		"groups":  groupCR.Spec.Members.Groups,
	})

	visitedGroups := make(map[string]struct{})
	allMembers, err := r.fetchUniqueGroupMembers(ctx, groupCR.Spec.GroupName, groupCR.Namespace, visitedGroups)
	if err != nil {
		r.log.WithError(err).Error("error fetching unique group members")
		return ctrl.Result{}, err
	}

	uniqueMembers := r.deduplicateMembers(allMembers)
	groupCR.Status.ReconciledUsers = uniqueMembers

	r.log.Info("fetching LDAP data for the users in the group")

	// Process LDAP data and update cache
	if err := r.processLDAPDataAndCache(ctx, uniqueMembers); err != nil {
		r.log.WithError(err).Error("error processing LDAP data and cache")
		return ctrl.Result{}, err
	}

	// Process all backends
	backendErrors := r.processAllBackends(ctx, groupCR, uniqueMembers)

	// Update status and handle errors
	return r.updateStatusAndHandleErrors(ctx, groupCR, backendErrors)
}

// processLDAPDataAndCache handles LDAP data fetching and cache operations
func (r *GroupReconciler) processLDAPDataAndCache(ctx context.Context, uniqueMembers []string) error {
	// Lock cache for read/write operations during reconciliation
	r.CacheMutex.Lock()
	defer r.CacheMutex.Unlock()

	r.log.Info("Acquired cache lock for group reconciliation operations")

	// Initialize LDAP user data map
	r.allLdapUserData = make(map[string]*structs.LDAPUser, len(uniqueMembers))

	// Get current user list from cache
	userList := r.getUserListFromCache(ctx)

	// Use a map to track unique UIDs to avoid duplicates
	uniqueUIDs := make(map[string]bool)
	for _, uid := range userList {
		uniqueUIDs[uid] = true
	}

	// Process each unique member
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
	}

	// Building a list of users who are active in LDAP.
	activeUserList := make([]string, 0, len(uniqueUIDs))
	for uid, isActive := range uniqueUIDs {
		if isActive {
			activeUserList = append(activeUserList, uid)
		}
	}

	// Update cache with new user list
	return r.updateUserListInCache(ctx, activeUserList)
}

// getUserListFromCache retrieves and unmarshals the user list from cache
func (r *GroupReconciler) getUserListFromCache(ctx context.Context) []string {
	userList := make([]string, 0)
	userListCache, err := r.Cache.Get(ctx, "user_list")
	if err != nil {
		r.log.WithError(err).Warn("user list not found in cache, will start with empty list")
		return userList
	}

	// Successfully retrieved from cache, now try to unmarshal
	r.log.WithField("user_list", userListCache).Debug("cached user list as JSON string")
	userListStr, ok := userListCache.(string)
	if !ok {
		r.log.Error("user list cache data is not a string, starting with empty list")
		return userList
	}

	if err := json.Unmarshal([]byte(userListStr), &userList); err != nil {
		r.log.WithError(err).Error("corrupted user list data in cache, invalidating and starting fresh")
		userList = make([]string, 0)

		// Invalidate the corrupted cache entry
		if deleteErr := r.Cache.Delete(ctx, "user_list"); deleteErr != nil {
			r.log.WithError(deleteErr).Warn("failed to delete corrupted user_list from cache")
		}
	}
	r.log.WithField("user_list_count", len(userList)).Info("user list retrieved from cache successfully")
	return userList
}

// updateUserListInCache marshals and stores the user list in cache
func (r *GroupReconciler) updateUserListInCache(ctx context.Context, userList []string) error {
	userListJSON, err := json.Marshal(userList)
	if err != nil {
		r.log.WithError(err).Error("error marshaling user list to JSON")
		return err
	}

	if err := r.Cache.Set(ctx, "user_list", string(userListJSON), cache.NoExpiration); err != nil {
		r.log.WithError(err).Error("error updating user list in cache")
		return err
	}

	return nil
}

// processAllBackends handles processing of all backends in the group CR
func (r *GroupReconciler) processAllBackends(
	ctx context.Context,
	groupCR *usernautdevv1alpha1.Group,
	uniqueMembers []string,
) map[string]string {
	backendErrors := make(map[string]string, 0)
	for _, backend := range groupCR.Spec.Backends {
		r.backendLogger = r.log.WithFields(logrus.Fields{
			"backend":      backend.Name,
			"backend_type": backend.Type,
		})

		if err := r.processSingleBackend(ctx, groupCR, backend, uniqueMembers); err != nil {
			r.backendLogger.WithError(err).Error("error processing backend")
			backendErrors[backend.Name] = err.Error()
		}
	}

	return backendErrors
}

// processSingleBackend handles processing of a single backend
func (r *GroupReconciler) processSingleBackend(ctx context.Context,
	groupCR *usernautdevv1alpha1.Group,
	backend usernautdevv1alpha1.Backend,
	uniqueMembers []string,
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
	teamID, err := r.fetchOrCreateTeam(ctx, groupCR.Spec.GroupName, backend.Name, backend.Type, backendClient)
	if err != nil {
		r.backendLogger.WithError(err).Error("error fetching or creating team")
		return err
	}
	r.backendLogger.WithField("team_id", teamID).Info("fetched or created team successfully")

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

	return nil
}

// updateStatusAndHandleErrors updates the CR status and handles any backend errors
func (r *GroupReconciler) updateStatusAndHandleErrors(ctx context.Context, groupCR *usernautdevv1alpha1.Group, backendErrors map[string]string) (ctrl.Result, error) {
	backendStatus := make([]usernautdevv1alpha1.BackendStatus, 0, len(groupCR.Spec.Backends))

	// Build status for each backend
	for _, backend := range groupCR.Spec.Backends {
		status := usernautdevv1alpha1.BackendStatus{
			Name: backend.Name,
			Type: backend.Type,
		}
		if msg, found := backendErrors[backend.Name]; found {
			status.Status = false
			status.Message = msg
		} else {
			status.Status = true
			status.Message = "Successful"
		}
		backendStatus = append(backendStatus, status)
	}

	// Update CR status
	groupCR.Status.BackendsStatus = backendStatus
	groupCR.UpdateStatus(false)
	if len(backendErrors) > 0 {
		groupCR.UpdateStatus(true)
	}
	if updateStatusErr := r.Status().Update(ctx, groupCR); updateStatusErr != nil {
		r.log.WithError(updateStatusErr).Error("error while updating final status")
	}

	// Return error if any backends failed
	if len(backendErrors) > 0 {
		return ctrl.Result{}, errors.New("failed to reconcile all backends")
	}

	return ctrl.Result{}, nil
}

// handleDeletion processes the deletion of a Group CR and its finalizer
func (r *GroupReconciler) handleDeletion(ctx context.Context, groupCR *usernautdevv1alpha1.Group) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(groupCR, groupFinalizer) {
		if err := r.deleteBackendsTeam(ctx, groupCR); err != nil {
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(groupCR, groupFinalizer)
		if err := r.Update(ctx, groupCR); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *GroupReconciler) deleteBackendsTeam(ctx context.Context, groupCR *usernautdevv1alpha1.Group) error {
	r.log.Info("Finalizer: starting Backends team deletion cleanup")

	for _, backend := range groupCR.Spec.Backends {
		transformed_group_name, err := utils.GetTransformedGroupName(r.AppConfig, backend.Type, groupCR.Spec.GroupName)
		backendLoggerInfo := r.log.WithFields(logrus.Fields{
			"team_name":             groupCR.Spec.GroupName,
			"transformed_team_name": transformed_group_name,
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

		teamDetailsMap := make(map[string]string)
		teamDetailsInCache, err := r.Cache.Get(ctx, transformed_group_name)
		if err == nil && teamDetailsInCache != "" {
			if jErr := json.Unmarshal([]byte(teamDetailsInCache.(string)), &teamDetailsMap); jErr != nil {
				backendLoggerInfo.WithError(err).Error("Finalizer: error unmarshalling team details from cache")
				return jErr
			}

			cacheKey := backend.Name + "_" + backend.Type

			if teamID, exists := teamDetailsMap[cacheKey]; exists && teamID != "" {
				backendLoggerInfo.Infof("Finalizer: Deleting team with (ID: %s) from Backend %s", teamID, backend.Type)

				if err := backendClient.DeleteTeamByID(ctx, teamID); err != nil {
					backendLoggerInfo.WithError(err).Error("Finalizer: failed to delete team from the backend")
					return err
				}
				backendLoggerInfo.Infof("Finalizer: Successfully deleted team with id '%s' from Backend %s", teamID, backend.Type)

				delete(teamDetailsMap, cacheKey)

				if err := r.Cache.Delete(ctx, transformed_group_name); err != nil {
					backendLoggerInfo.WithError(err).Error("Finalizer: failed to delete cache entry after cleanup")
					return err
				}

				if len(teamDetailsMap) > 0 {
					updatedCacheData, err := json.Marshal(teamDetailsMap)
					if err != nil {
						backendLoggerInfo.WithError(err).Error("Finalizer: failed to marshal updated team details for cache")
						return err
					}
					if err := r.Cache.Set(ctx, transformed_group_name, string(updatedCacheData), cache.NoExpiration); err != nil {
						backendLoggerInfo.WithError(err).Error("Finalizer: failed to update cache after deleting team")
						return err
					}
					backendLoggerInfo.Infof(
						"Finalizer: Updated cache after removing team ID '%s' for group '%s'", teamID, transformed_group_name)
				} else {
					backendLoggerInfo.Info("Finalizer: No more entries are there in the cache")
				}
			}
		}
	}
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

		userDetailsMap := make(map[string]string)
		userDetailsInCache, err := r.Cache.Get(ctx, userDetails.GetEmail())
		if err != nil && err != redis.Nil || userDetailsInCache == "" {
			r.backendLogger.WithError(err).Error("error fetching user details from cache")
			return nil, nil, err
		}

		userDetailsStr, ok := userDetailsInCache.(string)
		if !ok {
			r.backendLogger.WithField("user", user).Error("user details in cache are not of type string")
			return nil, nil, errors.New("user details in cache are not of type string")
		}

		if jErr := json.Unmarshal([]byte(userDetailsStr), &userDetailsMap); jErr != nil {
			r.backendLogger.WithField("user", user).WithError(jErr).Error("error unmarshalling user details from cache")
			return nil, nil, jErr
		}
		userID := userDetailsMap[backendName+"_"+backendType]
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

func (r *GroupReconciler) createUsersInBackendAndCache(ctx context.Context,
	users []string,
	backendName, backendType string,
	backendClient clients.Client) error {

	for _, user := range users {
		userDetails := r.allLdapUserData[user]
		if userDetails == nil {
			r.backendLogger.WithField("user", user).Warn("user not found in LDAP data, skipping user creation")
			continue
		}

		userDetailsMap := make(map[string]string)
		userDetailsInCache, err := r.Cache.Get(ctx, userDetails.GetEmail())
		if err == nil && userDetailsInCache != "" {
			// handle error for below statement
			if jErr := json.Unmarshal([]byte(userDetailsInCache.(string)), &userDetailsMap); jErr != nil {
				r.backendLogger.WithField("user", user).WithError(jErr).Error("error unmarshalling user details from cache")
				return jErr
			}
			userID := userDetailsMap[backendName+"_"+backendType]
			if userID != "" {
				r.backendLogger.WithField("user", user).Debug("user already exists in cache")
				continue
			}
		}

		// if user details are not found in cache, create a new user in backend
		newUser, err := backendClient.CreateUser(ctx, &structs.User{
			Email:     userDetails.GetEmail(),
			UserName:  user,
			Role:      fivetran.AccountReviewerRole,
			FirstName: userDetails.GetDisplayName(),
			LastName:  userDetails.GetSN(),
		})
		if err != nil {
			// TODO: handle the error in case user already exists in backend, we need to again populate the cache
			r.backendLogger.WithField("user", user).WithError(err).Error("error creating user in backend")
			return err
		}
		r.backendLogger.WithField("user", user).Info("created user in backend successfully")

		userDetailsMap[backendName+"_"+backendType] = newUser.ID
		toBeUpdated, _ := json.Marshal(userDetailsMap)
		if err := r.Cache.Set(ctx, userDetails.GetEmail(), string(toBeUpdated), cache.NoExpiration); err != nil {
			r.backendLogger.Error(err, "error updating user details in cache")
			return err
		}
		r.backendLogger.WithField("user", user).Info("updated user details in cache successfully")
	}
	return nil
}

func (r *GroupReconciler) fetchOrCreateTeam(ctx context.Context,
	groupName string,
	backendName, backendType string,
	backendClient clients.Client) (string, error) {

	// transforming the group name
	transformed_group_name, err := utils.GetTransformedGroupName(r.AppConfig, backendType, groupName)
	if err != nil {
		r.backendLogger.WithError(err).Error("error transforming the group Name")
		return "", err
	}

	teamDetailsMap := make(map[string]string)

	teamDetailsInCache, err := r.Cache.Get(ctx, transformed_group_name)
	if err == nil && teamDetailsInCache != "" {
		if jErr := json.Unmarshal([]byte(teamDetailsInCache.(string)), &teamDetailsMap); jErr != nil {
			r.backendLogger.WithError(jErr).Error("error unmarshalling team details from cache")
			return "", jErr
		}
		// Check if the team details for the backend exist in cache
		if teamID, exists := teamDetailsMap[backendName+"_"+backendType]; exists && teamID != "" {
			r.backendLogger.WithField("teamID", teamID).Info("team details found in cache")
			return teamID, nil
		}
	}
	// If team details are not found in cache, create a new team
	r.backendLogger.Info("team details not found in cache, creating a new team")

	newTeam, err := backendClient.CreateTeam(ctx, &structs.Team{
		Name:        transformed_group_name,
		Description: "team for " + groupName,
		Role:        fivetran.AccountReviewerRole,
	})
	if err != nil {
		// TODO: handle the error in case team already exists in backend, we need to again populate the cache
		r.backendLogger.WithError(err).Error("error creating team in backend")
		return "", err
	}

	r.backendLogger.Info("created team in backend successfully")

	// Create the team in cache
	teamDetailsMap[backendName+"_"+backendType] = newTeam.ID
	toBeUpdated, _ := json.Marshal(teamDetailsMap)
	if err := r.Cache.Set(ctx, transformed_group_name, string(toBeUpdated), cache.NoExpiration); err != nil {
		r.backendLogger.WithError(err).Error("error updating team details in cache")
		return "", err
	}

	r.backendLogger.Info("updated team details in cache successfully")

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

	return ctrl.NewControllerManagedBy(mgr).
		For(&usernautdevv1alpha1.Group{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Watches(
			client.Object(&usernautdevv1alpha1.Group{}),
			handler.EnqueueRequestsFromMapFunc(mapFunc),
		).
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

		transformedGroupName, err := utils.GetTransformedGroupName(r.AppConfig, dependsOn.Type, groupName)
		if err != nil {
			r.backendLogger.WithError(err).Error("error transforming the group Name")
			return false, err
		}
		err = r.ldapDependantChecks(dependsOn, transformedGroupName)
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
	// Check if the RoverGroup exists or not by checking the cache...
	teamDetailsInCache, err := r.Cache.Get(context.Background(), groupName)
	if err != nil || teamDetailsInCache == "" {
		r.backendLogger.WithError(err).Error("rover team details not found in cache, skipping ldap sync")
		return err
	}
	return nil
}

func isGroupCRHasDependants(backends []usernautdevv1alpha1.Backend, dependsOn config.Dependant) bool {
	for _, backend := range backends {
		if backend.Type == dependsOn.Type && backend.Name == dependsOn.Name {
			return true
		}
	}
	return false
}
