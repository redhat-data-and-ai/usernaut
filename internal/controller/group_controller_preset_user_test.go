package controller

import (
	"context"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/redhat-data-and-ai/usernaut/pkg/cache"
	"github.com/redhat-data-and-ai/usernaut/pkg/cache/inmemory"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/store"
)

type recordingBackendClient struct {
	clients.Client
	created []*structs.User
}

func (c *recordingBackendClient) CreateUser(_ context.Context, user *structs.User) (*structs.User, error) {
	copied := *user
	c.created = append(c.created, &copied)
	return &structs.User{
		ID:       "backend-id-" + user.UserName,
		Email:    user.Email,
		UserName: user.UserName,
	}, nil
}

func TestCreateUsersInBackendAndCache_userNameByBackend(t *testing.T) {
	memCache, err := cache.New(&cache.Config{
		Driver: cache.DriverMemory,
		InMemory: &inmemory.Config{
			DefaultExpiration: int32(-1),
			CleanupInterval:   int32(-1),
		},
	})
	require.NoError(t, err)

	ldapUsers := map[string]*structs.LDAPUser{
		"ctolosa": {
			Email:    "ctolosa@redhat.com",
			UID:      "ctolosa",
			RhatUUID: "11111111-2222-3333-4444-555555555555",
		},
	}
	reconciler := &GroupReconciler{
		Store:      store.New(memCache),
		CacheMutex: &sync.RWMutex{},
	}
	log := logrus.NewEntry(logrus.New())

	presetClient := &recordingBackendClient{}
	err = reconciler.createUsersInBackendAndCache(
		context.Background(),
		[]string{"ctolosa"},
		"preset",
		"preset",
		presetClient,
		log,
		ldapUsers,
	)
	require.NoError(t, err)
	require.Len(t, presetClient.created, 1)
	require.Equal(t, "ctolosa", presetClient.created[0].UserName)
	require.Equal(t, "11111111-2222-3333-4444-555555555555", presetClient.created[0].RhatUUID)
	require.Equal(t, "ctolosa@redhat.com", presetClient.created[0].Email)

	fivetranClient := &recordingBackendClient{}
	err = reconciler.createUsersInBackendAndCache(
		context.Background(),
		[]string{"ctolosa"},
		"fivetran",
		"fivetran",
		fivetranClient,
		log,
		ldapUsers,
	)
	require.NoError(t, err)
	require.Len(t, fivetranClient.created, 1)
	require.Equal(t, "ctolosa", fivetranClient.created[0].UserName)
}
