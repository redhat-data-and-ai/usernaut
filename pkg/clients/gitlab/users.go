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

package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	"github.com/sirupsen/logrus"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	log := logger.Logger(ctx).WithField("service", "gitlab")
	log.Info("fetching all users")

	userEmailMap := make(map[string]*structs.User)
	userIDMap := make(map[string]*structs.User)

	if g.dependantExists {
		// Since users will be fetched using Rover as LDAP, we don't need to fetch users from Gitlab API
		// if it has dependant in gitlab backend configs
		return userEmailMap, userIDMap, nil
	}

	human := true
	active := true
	opt := &gitlab.ListUsersOptions{
		Humans: &human,
		Active: &active,
		ListOptions: gitlab.ListOptions{
			PerPage: 100, // Maximum allowed
			Page:    1,
		},
	}

	for {
		users, resp, err := g.gitlabClient.Users.ListUsers(opt)
		if err != nil {
			return nil, nil, err
		}

		for _, user := range users {
			userEmailMap[user.Email] = userDetails(user)
			userIDMap[fmt.Sprintf("%d", user.ID)] = userDetails(user)
		}

		// Check if we got fewer users than requested (last page)
		if len(users) < opt.PerPage {
			break
		}

		// For offset pagination, check NextPage
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	log.WithField("total_user_count", len(userIDMap)).Info("found users")
	return userEmailMap, userIDMap, nil
}

func (g *GitlabClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"userID":  userID,
	})
	log.Info("fetching user details")
	var user *gitlab.User

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		var numErr *strconv.NumError
		// If the error is a syntax error, it's not a number, so we assume it's a username.
		// Other errors, like a range error for a number that's too large, are treated as invalid input.
		if errors.As(err, &numErr) && numErr.Err == strconv.ErrSyntax {
			log.Info("userID is not an integer, assuming it is a username")
			users, resp, listErr := g.gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{
				Username: &userID,
			})
			if listErr != nil {
				if resp != nil && resp.StatusCode == http.StatusNotFound {
					return nil, fmt.Errorf("user %s not found in gitlab (404)", userID)
				}
				log.WithError(listErr).Error("Failed to fetch existing user")
				return nil, listErr
			}
			if len(users) > 0 && resp.StatusCode == http.StatusOK {
				log.Infof("found user %s details in gitlab backend", userID)
				return userDetails(users[0]), nil
			} else {
				// this handles the case where user never logged in to gitlab
				// so user details like userID is not found in gitlab
				// TODO: need to handle the case when the user login in gitlab
				log.Warnf("unable to find user %s details in gitlab backend", userID)
				return &structs.User{
					ID:       userID,
					UserName: userID,
				}, nil
			}
		} else {
			// The userID is not a valid username-like string or is a number out of range.
			return nil, fmt.Errorf("invalid userID format: %w", err)
		}
	} else {
		// If userID is an integer, fetch user by ID
		var getErr error
		user, _, getErr = g.gitlabClient.Users.GetUser(userIDInt, gitlab.GetUsersOptions{})
		if getErr != nil {
			return nil, getErr
		}
		log.Info("found user details")
		return userDetails(user), nil
	}
}

func (g *GitlabClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"user":    u,
	})
	log.Info("creating user")

	if g.ldapSync {
		user, err := g.FetchUserDetails(ctx, u.UserName)
		if err != nil {
			log.WithError(err).Error("Failed to fetch user details")
			return nil, err
		}
		if user.Email == "" {
			user.Email = u.Email
		}
		return user, nil
	}

	// Use Gitlab SDK to create a user
	resetPassword := true // Still required by GitLab API, but user will auth via LDAP
	createUserOptions := &gitlab.CreateUserOptions{
		Email:         &u.Email,
		Username:      &u.UserName,
		Name:          &u.FirstName,
		ResetPassword: &resetPassword, // Required by API, but unused with LDAP
	}
	user, resp, err := g.gitlabClient.Users.CreateUser(createUserOptions)
	if err != nil {
		if resp.StatusCode == http.StatusForbidden {
			log.WithError(err).Error(
				"user creation forbidden, check ldapSync for gitlab backend or obtain admin privileges",
			)
			return nil, err
		}
		return nil, err
	}

	return userDetails(user), nil
}

func (g *GitlabClient) DeleteUser(ctx context.Context, userID string) error {
	log := logger.Logger(ctx).WithFields(logrus.Fields{
		"service": "gitlab",
		"userID":  userID,
	})
	log.Info("deleting user")

	if g.ldapSync {
		return nil
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.WithError(err).Error("Failed to convert userID to int")
		return err
	}
	_, err = g.gitlabClient.Users.DeleteUser(userIDInt)
	if err != nil {
		log.WithError(err).Error("Failed to delete user")
		return err
	}
	log.Info("user deleted successfully")
	return err
}

func userDetails(u *gitlab.User) *structs.User {
	return &structs.User{
		ID:          fmt.Sprintf("%d", u.ID),
		Email:       u.Email,
		UserName:    u.Username,
		DisplayName: u.Name,
	}
}
