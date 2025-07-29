package atlan

import (
	"context"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
)

func (ac *AtlanClient) FetchAllUsers(ctx context.Context) (map[string]*structs.User, map[string]*structs.User, error) {
	// Users are synced via LDAP, return empty maps
	return make(map[string]*structs.User), make(map[string]*structs.User), nil
}

func (ac *AtlanClient) FetchUserDetails(ctx context.Context, userID string) (*structs.User, error) {
	// Users are synced via LDAP, return empty user
	return &structs.User{}, nil
}

func (ac *AtlanClient) CreateUser(ctx context.Context, u *structs.User) (*structs.User, error) {
	// Users are synced via LDAP, return minimal user struct
	return &structs.User{
		ID: u.UserName,
	}, nil
}

func (ac *AtlanClient) DeleteUser(ctx context.Context, userID string) error {
	// Users are synced via LDAP, return nil
	return nil
}
