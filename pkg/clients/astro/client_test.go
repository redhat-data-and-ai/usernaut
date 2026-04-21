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

package astro

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestClient creates an AstroClient with a mock server
func createTestClient(t *testing.T, handler http.Handler) (*AstroClient, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)

	client := &AstroClient{
		config: &AstroConfig{
			APIToken:       "test-token",
			OrganizationID: "test-org-id",
			BaseURL:        server.URL,
		},
		client: server.Client(),
	}

	return client, server
}

func TestNewClient(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		connection := map[string]interface{}{
			"api_token":       "test-token",
			"organization_id": "test-org-id",
		}
		poolCfg := testPoolConfig()
		hystrixCfg := testHystrixConfig()

		client, err := NewClient(connection, poolCfg, hystrixCfg)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "test-token", client.config.APIToken)
		assert.Equal(t, "test-org-id", client.config.OrganizationID)
		assert.Equal(t, DefaultBaseURL, client.config.BaseURL)
	})

	t.Run("CustomBaseURL", func(t *testing.T) {
		connection := map[string]interface{}{
			"api_token":       "test-token",
			"organization_id": "test-org-id",
			"base_url":        "https://custom.api.com",
		}
		poolCfg := testPoolConfig()
		hystrixCfg := testHystrixConfig()

		client, err := NewClient(connection, poolCfg, hystrixCfg)
		require.NoError(t, err)
		assert.Equal(t, "https://custom.api.com", client.config.BaseURL)
	})

	t.Run("MissingAPIToken", func(t *testing.T) {
		connection := map[string]interface{}{
			"organization_id": "test-org-id",
		}
		poolCfg := testPoolConfig()
		hystrixCfg := testHystrixConfig()

		_, err := NewClient(connection, poolCfg, hystrixCfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api_token")
	})

	t.Run("MissingOrgID", func(t *testing.T) {
		connection := map[string]interface{}{
			"api_token": "test-token",
		}
		poolCfg := testPoolConfig()
		hystrixCfg := testHystrixConfig()

		_, err := NewClient(connection, poolCfg, hystrixCfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization_id")
	})
}

func TestFetchAllUsers(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		usersResp := AstroUsersResponse{
			Users: []AstroUser{
				{
					ID:               "user-1",
					Username:         "user1@redhat.com",
					FullName:         "User One",
					Status:           "ACTIVE",
					OrganizationRole: "ORGANIZATION_MEMBER",
				},
				{
					ID:               "user-2",
					Username:         "user2@redhat.com",
					FullName:         "User Two",
					Status:           "ACTIVE",
					OrganizationRole: "ORGANIZATION_MEMBER",
				},
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "/v1/organizations/test-org-id/users")
			assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(usersResp)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		usersByID, usersByEmail, err := client.FetchAllUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, usersByID, 2)
		assert.Len(t, usersByEmail, 2)

		assert.Equal(t, "user-1", usersByID["user-1"].ID)
		assert.Equal(t, "user1@redhat.com", usersByID["user-1"].Email)
		assert.Equal(t, "User", usersByID["user-1"].FirstName)
		assert.Equal(t, "One", usersByID["user-1"].LastName)

		assert.Equal(t, "user-2", usersByEmail["user2@redhat.com"].ID)
	})

	t.Run("EmptyResponse", func(t *testing.T) {
		usersResp := AstroUsersResponse{Users: []AstroUser{}}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(usersResp)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		usersByID, usersByEmail, err := client.FetchAllUsers(ctx)
		require.NoError(t, err)
		assert.Len(t, usersByID, 0)
		assert.Len(t, usersByEmail, 0)
	})
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		inviteResp := CreateInviteResponse{
			UserID: "new-user-id",
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "/invites")

			var req CreateInviteRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "newuser@redhat.com", req.InviteeEmail)
			assert.Equal(t, DefaultOrganizationRole, req.Role)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(inviteResp)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		user := &structs.User{
			Email:    "newuser@redhat.com",
			UserName: "newuser@redhat.com",
		}

		created, err := client.CreateUser(ctx, user)
		require.NoError(t, err)
		assert.Equal(t, "new-user-id", created.ID)
		assert.Equal(t, "newuser@redhat.com", created.Email)
	})

	t.Run("MissingEmail", func(t *testing.T) {
		client, server := createTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()

		user := &structs.User{UserName: "testuser"}
		_, err := client.CreateUser(ctx, user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "/users/user-to-delete/roles")

			var req UpdateUserRoleRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Nil(t, req.OrganizationRole)

			w.WriteHeader(http.StatusOK)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.DeleteUser(ctx, "user-to-delete")
		require.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error": "User not found"}`))
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.DeleteUser(ctx, "nonexistent-user")
		assert.Error(t, err)
	})
}

func TestFetchAllTeams(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		teamsResp := AstroTeamsResponse{
			Teams: []AstroTeam{
				{
					ID:               "team-1",
					Name:             "dataverse-aggregate-test",
					Description:      "Test team",
					OrganizationRole: "ORGANIZATION_MEMBER",
				},
				{
					ID:               "team-2",
					Name:             "dataverse-aggregate-prod",
					Description:      "Prod team",
					OrganizationRole: "ORGANIZATION_MEMBER",
				},
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "/teams")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(teamsResp)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		teams, err := client.FetchAllTeams(ctx)
		require.NoError(t, err)
		assert.Len(t, teams, 2)

		assert.Equal(t, "dataverse-aggregate-test", teams["team-1"].Name)
		assert.Equal(t, "Test team", teams["team-1"].Description)
	})
}

func TestCreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		createdTeam := AstroTeam{
			ID:               "new-team-id",
			Name:             "new-team",
			Description:      "New team description",
			OrganizationRole: "ORGANIZATION_MEMBER",
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "/teams")

			var req CreateTeamRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "new-team", req.Name)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(createdTeam)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		team := &structs.Team{
			Name:        "new-team",
			Description: "New team description",
		}

		created, err := client.CreateTeam(ctx, team)
		require.NoError(t, err)
		assert.Equal(t, "new-team-id", created.ID)
		assert.Equal(t, "new-team", created.Name)
	})
}

func TestDeleteTeamByID(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Contains(t, r.URL.Path, "/teams/team-to-delete")

			w.WriteHeader(http.StatusNoContent)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.DeleteTeamByID(ctx, "team-to-delete")
		require.NoError(t, err)
	})
}

func TestFetchTeamMembersByTeamID(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		membersResp := AstroTeamMembersResponse{
			TeamMembers: []AstroTeamMember{
				{
					UserID:   "user-1",
					Username: "user1@redhat.com",
					FullName: "User One",
				},
				{
					UserID:   "user-2",
					Username: "user2@redhat.com",
					FullName: "User Two",
				},
			},
		}

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Contains(t, r.URL.Path, "/teams/team-1/members")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(membersResp)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		members, err := client.FetchTeamMembersByTeamID(ctx, "team-1")
		require.NoError(t, err)
		assert.Len(t, members, 2)

		assert.Equal(t, "user-1", members["user-1"].ID)
		assert.Equal(t, "user1@redhat.com", members["user-1"].Email)
	})
}

func TestAddUserToTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Contains(t, r.URL.Path, "/teams/team-1/members")

			var req AddTeamMembersRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Contains(t, req.MemberIDs, "user-1")
			assert.Contains(t, req.MemberIDs, "user-2")

			w.WriteHeader(http.StatusOK)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.AddUserToTeam(ctx, "team-1", []string{"user-1", "user-2"})
		require.NoError(t, err)
	})

	t.Run("EmptyUserList", func(t *testing.T) {
		client, server := createTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make request for empty user list")
		}))
		defer server.Close()

		err := client.AddUserToTeam(ctx, "team-1", []string{})
		require.NoError(t, err)
	})
}

func TestRemoveUserFromTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			callCount++
			w.WriteHeader(http.StatusNoContent)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.RemoveUserFromTeam(ctx, "team-1", []string{"user-1", "user-2"})
		require.NoError(t, err)
		assert.Equal(t, 2, callCount) // Should make 2 calls, one per user
	})

	t.Run("UserNotFoundIsOK", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		client, server := createTestClient(t, handler)
		defer server.Close()

		err := client.RemoveUserFromTeam(ctx, "team-1", []string{"nonexistent-user"})
		require.NoError(t, err)
	})
}

func TestReconcileGroupParams(t *testing.T) {
	ctx := context.Background()

	t.Run("NoOp", func(t *testing.T) {
		client, server := createTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("Should not make any request")
		}))
		defer server.Close()

		params := structs.TeamParams{
			Property: "test",
			Value:    []string{"value1", "value2"},
		}

		err := client.ReconcileGroupParams(ctx, "team-1", params)
		require.NoError(t, err)
	})
}

func TestGetConfig(t *testing.T) {
	client := &AstroClient{
		config: &AstroConfig{
			APIToken:       "test-token",
			OrganizationID: "test-org",
			BaseURL:        "https://api.test.com",
		},
	}

	config := client.GetConfig()
	assert.Equal(t, "test-token", config.APIToken)
	assert.Equal(t, "test-org", config.OrganizationID)
	assert.Equal(t, "https://api.test.com", config.BaseURL)
}

// Helper functions for test configuration
func testPoolConfig() httpclient.ConnectionPoolConfig {
	return httpclient.ConnectionPoolConfig{
		Timeout:            10000,
		KeepAliveTimeout:   60000,
		MaxIdleConnections: 10,
	}
}

func testHystrixConfig() httpclient.HystrixResiliencyConfig {
	return httpclient.HystrixResiliencyConfig{
		MaxConcurrentRequests:     100,
		RequestVolumeThreshold:    20,
		CircuitBreakerSleepWindow: 5000,
		ErrorPercentThreshold:     50,
		CircuitBreakerTimeout:     10000,
	}
}
