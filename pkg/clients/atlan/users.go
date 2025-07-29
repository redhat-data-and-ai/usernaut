package atlan

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

func (ac *AtlanClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	// Users are synced via LDAP, return empty maps like Rover does
	return make(map[string]*structs.User), make(map[string]*structs.User), nil
}

func (ac *AtlanClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	// Users are synced via LDAP, return empty user like Rover does
	return &structs.User{}, nil
}

func (ac *AtlanClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	// Users are synced via LDAP, return minimal user struct like Rover does
	return &structs.User{
		ID: u.UserName,
	}, nil
}

func (ac *AtlanClient) DeleteUser(ctx context.Context, userID string) error {
	// Users are synced via LDAP, return nil like Rover does
	return nil
}
