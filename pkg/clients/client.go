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

package clients

import (
	"context"
	"errors"
	"strings"

	"github.com/redhat-data-and-ai/usernaut/pkg/clients/fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

var (
	// ErrInvalidBackend is returned when an invalid backend type is provided
	ErrInvalidBackend = errors.New("invalid backend")
)

type Client interface {
	// Fetches all the users onboarded over the platform
	// returns 2 maps where:
	// 1st map will have ID as key in order to map with team membership response
	// and 2nd will have email as key
	FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error)
	// Fetches user details based on unique userID
	FetchUserDetails(ctx context.Context, userID string) (*structs.User, error)
	// Onboards the user on the backend
	CreateUser(ctx context.Context, u *structs.User) (*structs.User, error)
	// Drop User from the backend
	DeleteUser(ctx context.Context, userID string) error

	// Fetches all the teams on the backend
	FetchAllTeams(ctx context.Context) (map[string]structs.Team, error)
	// Fetch team details by ID or unique key
	FetchTeamDetails(ctx context.Context, teamID string) (*structs.Team, error)
	// Create a new team/role
	CreateTeam(ctx context.Context, team *structs.Team) (*structs.Team, error)
	// Drop the team from respective backend
	DeleteTeamByID(ctx context.Context, teamID string) error

	// Returns the list of users present under a team
	FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error)
	// Adds a member to the team
	AddUserToTeam(ctx context.Context, teamID, userID string) error
	// Removes a member from the team
	RemoveUserFromTeam(ctx context.Context, teamID, userID string) error
}

type Backends struct {
	// Snowflake map[string]Snowflake
	Fivetran map[string]*fivetran.FivetranConfig
}

func (backends *Backends) New(name, backendType string) (Client, error) {
	switch strings.ToLower(backendType) {
	case "fivetran":
		fivetranConfig := backends.Fivetran[name]
		if fivetranConfig == nil {
			return nil, errors.New("fivetran client isn't configured")
		}
		return fivetran.NewClient(fivetranConfig.ApiKey, fivetranConfig.ApiSecret), nil
	default:
		return nil, ErrInvalidBackend
	}
}
