package atlan

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

func (ac *AtlanClient) FetchTeamMembersByTeamID(ctx context.Context, teamID string) (map[string]*structs.User, error) {
	// Team membership is synced via LDAP, returning empty map
	return make(map[string]*structs.User), nil
}

func (ac *AtlanClient) AddUserToTeam(ctx context.Context, teamID, userID string) error {
	// Team membership is synced via LDAP, returning nil
	return nil
}

func (ac *AtlanClient) RemoveUserFromTeam(ctx context.Context, teamID, userID string) error {
	// Team membership is synced via LDAP, returning nil
	return nil
}
