package fivetran

import (
	"context"

	"github.com/fivetran/go-fivetran"
	"github.com/fivetran/go-fivetran/teams"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

type FivetranClient struct {
	fivetranClient *fivetran.Client
}

type IFivetran interface {
	// Operation on Team
	CreateTeam(context.Context, *Team) (*Team, error)
	UpdateTeam(context.Context, *UpdateTeam) (*Team, error)
	DeleteTeamByID(context.Context, string) error
	FetchAllTeams(context.Context) (map[string]teams.TeamData, error)
	FetchTeamDetails(context.Context, string) (teams.TeamData, error)

	// Operations on Team membership
	FetchTeamMembersByTeamID(context.Context, string, map[string]*structs.User) (map[string]*structs.User, error)
	AddUserToTeam(context.Context, string, string) (teams.TeamUserMembership, error)
	RemoveUserFromTeam(context.Context, string, string) error

	// Operation on Users
	InviteUser(context.Context, *structs.User) (*structs.User, error)
	UpdateUser(context.Context, *structs.User) (*structs.User, error)
	DeleteUser(context.Context, string) error
	FetchAllUsers(context.Context) (map[string]*structs.User, map[string]*structs.User, error)
	FetchUserDetails(context.Context, string) (*structs.User, error)
}

func NewClient(apiKey, apiSecret string) IFivetran {
	return &FivetranClient{
		fivetranClient: fivetran.New(apiKey, apiSecret),
	}
}
