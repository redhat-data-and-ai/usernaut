package clients

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
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
