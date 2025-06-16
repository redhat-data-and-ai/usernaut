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
	"slices"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/go-redis/redis"
	usernautdevv1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	"github.com/sirupsen/logrus"
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
}

// +kubebuilder:rbac:groups=usernaut.dev,resources=groups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=usernaut.dev,resources=groups/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=usernaut.dev,resources=groups/finalizers,verbs=update

func (r *GroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = logger.WithRequestId(ctx, controller.ReconcileIDFromContext(ctx))
	r.log = logger.Logger(ctx).WithFields(logrus.Fields{
		"request": req.NamespacedName.String(),
	})

	groupCR := &usernautdevv1alpha1.Group{}
	if err := r.Client.Get(ctx, req.NamespacedName, groupCR); err != nil {
		r.log.WithError(err).Error("error fetching the group CR")
		return ctrl.Result{}, err
	}

	r.log = logger.Logger(ctx).WithFields(logrus.Fields{
		"request": req.NamespacedName.String(),
		"group":   groupCR.Spec.GroupName,
		"members": len(groupCR.Spec.Members),
	})

	r.log.Info("reconciling the group against the backends")

	// fetch all the data from LDAP for the users in the group
	r.allLdapUserData = make(map[string]*structs.LDAPUser, 0)
	for _, user := range groupCR.Spec.Members {
		ldapUserData, err := r.LdapConn.GetUserLDAPData(ctx, user)
		if err != nil {
			r.log.WithError(err).Error("error fetching user data from LDAP")
			continue
		}

		ldapUser := &structs.LDAPUser{}
		err = utils.MapToStruct(ldapUserData, ldapUser)
		if err != nil {
			r.log.WithError(err).Error("error converting LDAP user data to struct")
			continue
		}

		r.allLdapUserData[ldapUser.UID] = ldapUser
	}

	for _, backend := range groupCR.Spec.Backends {

		r.backendLogger = r.log.WithFields(logrus.Fields{
			"backend":      backend.Name,
			"backend_type": backend.Type,
		})

		// process each backend in the group CR
		backendClient, err := clients.New(backend.Name, backend.Type, r.AppConfig.BackendMap)
		if err != nil {
			r.backendLogger.WithError(err).Error("error creating backend client")
			return ctrl.Result{}, err
		}
		r.backendLogger.Debug("created backend client successfully")

		// fetch the teamID or create a new team if it doesn't exist
		teamID, err := r.fetchOrCreateTeam(ctx, groupCR.Spec.GroupName, backend.Name, backend.Type, backendClient)
		if err != nil {
			r.backendLogger.WithError(err).Error("error fetching or creating team")
			return ctrl.Result{}, err
		}
		r.backendLogger.WithField("team_id", teamID).Info("fetched or created team successfully")

		// create the users in backend and cache if they don't exist
		err = r.createUsersInBackendAndCache(ctx, groupCR.Spec.Members, backend.Name, backend.Type, backendClient)
		if err != nil {
			r.backendLogger.WithError(err).Error("error creating users in backend and cache")
			return ctrl.Result{}, err
		}
		r.backendLogger.Info("created users in backend and cache successfully")

		// fetch the existing team members in the backend
		members, err := backendClient.FetchTeamMembersByTeamID(ctx, teamID)
		if err != nil {
			r.backendLogger.WithError(err).Error("error fetching team members")
			return ctrl.Result{}, err
		}

		// members field doesn't contains an email mapped to the user, we need to map it before finding the diff
		r.backendLogger.WithField("team_members_count", len(members)).Info("fetched team members successfully")

		usersToAdd, usersToRemove, err := r.processUsers(ctx, groupCR.Spec.Members, members, backend.Name, backend.Type)
		if err != nil {
			r.backendLogger.WithError(err).Error("error processing users")
			return ctrl.Result{}, err
		}

		r.backendLogger.Info("reconciling users for the team")
		for _, userID := range usersToAdd {
			if err := backendClient.AddUserToTeam(ctx, teamID, userID); err != nil {
				r.backendLogger.WithField("user_id", userID).WithError(err).Error("error adding user to team")
				return ctrl.Result{}, err
			}
		}
		r.backendLogger.WithField("users_to_add", usersToAdd).Info("added users to team successfully")

		for _, userID := range usersToRemove {
			if err := backendClient.RemoveUserFromTeam(ctx, teamID, userID); err != nil {
				r.backendLogger.WithField("user_id", userID).WithError(err).Error("error removing user from team")
				return ctrl.Result{}, err
			}
		}
		r.backendLogger.WithField("users_to_remove", usersToRemove).Info("removed users from team successfully")
	}

	return ctrl.Result{}, nil
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
	for userID, _ := range existingTeamMembers {
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

	teamDetailsMap := make(map[string]string)

	teamDetailsInCache, err := r.Cache.Get(ctx, groupName)
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
		Name:        groupName,
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
	if err := r.Cache.Set(ctx, groupName, string(toBeUpdated), cache.NoExpiration); err != nil {
		r.backendLogger.WithError(err).Error("error updating team details in cache")
		return "", err
	}

	r.backendLogger.Info("updated team details in cache successfully")

	return newTeam.ID, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usernautdevv1alpha1.Group{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
