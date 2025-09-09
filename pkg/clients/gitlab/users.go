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
	"fmt"
	"strconv"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func (g *GitlabClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	users, _, err := g.gitlabClient.Users.ListUsers(nil)
	if err != nil {
		return nil, nil, err
	}
	userEmailMap := make(map[string]*structs.User)
	userIDMap := make(map[string]*structs.User)
	for _, user := range users {
		userEmailMap[user.Email] = userDetails(user)
		userIDMap[fmt.Sprintf("%d", user.ID)] = userDetails(user)
	}
	return userEmailMap, userIDMap, nil
}

func (g *GitlabClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, err
	}
	user, _, err := g.gitlabClient.Users.GetUser(userIDInt, gitlab.GetUsersOptions{})
	if err != nil {
		return nil, err
	}
	return userDetails(user), nil
}

func (g *GitlabClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	log := logger.Logger(ctx)
	log.Info("Create GitLab user")

	if g.ldapSync {
		users, _, fetchErr := g.gitlabClient.Users.ListUsers(&gitlab.ListUsersOptions{
			Username: &u.UserName,
		})
		if fetchErr == nil && len(users) > 0 {
			return userDetails(users[0]), nil
		}
		log.WithField("fetch_error", fetchErr).Error("Failed to fetch existing user")
		return nil, fetchErr
	}

	// Use Gitlab SDK to create a user
	resetPassword := true // Still required by GitLab API, but user will auth via LDAP
	createUserOptions := &gitlab.CreateUserOptions{
		Email:         &u.Email,
		Username:      &u.UserName,
		Name:          &u.FirstName,
		ResetPassword: &resetPassword, // Required by API, but unused with LDAP
	}
	user, _, err := g.gitlabClient.Users.CreateUser(createUserOptions)
	if err != nil {
		return nil, err
	}

	return userDetails(user), nil
}

func (g *GitlabClient) DeleteUser(ctx context.Context, userID string) error {
	// TODO: Implement using g.client.Users.DeleteUser
	return nil
}

func userDetails(u *gitlab.User) *structs.User {
	return &structs.User{
		ID:       fmt.Sprintf("%d", u.ID),
		Email:    u.Email,
		UserName: u.Username,
	}
}
