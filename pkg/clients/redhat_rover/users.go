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

package redhatrover

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

func (rC *RoverClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	// this doesn't need any implementation as Rover is the LDAP
	return make(map[string]*structs.User), make(map[string]*structs.User), nil
}

func (rC *RoverClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	// this doesn't need any implementation as Rover is the LDAP
	return &structs.User{}, nil
}

func (rC *RoverClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	// as rover is the LDAP, no need to create user here
	// field UserName is used as ID in Rover
	return &structs.User{
		ID: u.UserName,
	}, nil
}

func (rC *RoverClient) DeleteUser(ctx context.Context, userID string) error {
	// this doesn't need any implementation as Rover is the LDAP
	return nil
}
