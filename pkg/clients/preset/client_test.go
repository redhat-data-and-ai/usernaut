package preset

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTeamSlug = "probe-team"

func testSCIMPath(suffix string) string {
	return "/api/v1/teams/" + testTeamSlug + "/scim/v2" + suffix
}

func testHTTPConfigs() (httpclient.ConnectionPoolConfig, httpclient.HystrixResiliencyConfig) {
	return httpclient.ConnectionPoolConfig{
			Timeout:            5000,
			KeepAliveTimeout:   600000,
			MaxIdleConnections: 10,
		},
		httpclient.HystrixResiliencyConfig{
			MaxConcurrentRequests:     100,
			RequestVolumeThreshold:    100,
			CircuitBreakerSleepWindow: 5000,
			ErrorPercentThreshold:     100,
			CircuitBreakerTimeout:     30000,
		}
}

func newTestPresetClient(t *testing.T, baseURL string) *PresetClient {
	t.Helper()

	poolCfg, hystrixCfg := testHTTPConfigs()
	client, err := httpclient.InitializeClient(
		"preset-test",
		poolCfg,
		hystrixCfg,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(0, 0)),
		1,
		nil,
	)
	require.NoError(t, err)

	return &PresetClient{
		client:    client,
		baseURL:   baseURL,
		scimToken: "test-token",
		teamSlug:  testTeamSlug,
	}
}

func makeUserIDs(count int) []string {
	ids := make([]string, count)
	for i := range ids {
		ids[i] = fmt.Sprintf("user-%d", i)
	}
	return ids
}

func decodePatchRequest(t *testing.T, r *http.Request) scimPatchRequest {
	t.Helper()
	var patchReq scimPatchRequest
	err := json.NewDecoder(r.Body).Decode(&patchReq)
	require.NoError(t, err)
	return patchReq
}

func addBatchMemberCount(patchReq scimPatchRequest) int {
	if len(patchReq.Operations) != 1 {
		return 0
	}
	members, ok := patchReq.Operations[0].Value.([]interface{})
	if !ok {
		return 0
	}
	return len(members)
}

func TestResponseStatusFromAPIError(t *testing.T) {
	err := &apiError{StatusCode: http.StatusConflict, Body: []byte("conflict")}
	assert.Equal(t, http.StatusConflict, responseStatus(err))
	assert.True(t, isResponseStatus(err, http.StatusConflict))
	assert.False(t, isResponseStatus(err, http.StatusNotFound))
	assert.Equal(t, 0, responseStatus(fmt.Errorf("other error")))
}

func TestRateLimitBackoff(t *testing.T) {
	assert.Equal(t, presetRateLimitDefaultBackoff, rateLimitBackoff(http.Header{}))
	assert.Equal(t, 2*time.Second, rateLimitBackoff(http.Header{"Retry-After": []string{"2"}}))
	assert.Equal(t, time.Duration(0), rateLimitBackoff(http.Header{"Retry-After": []string{"0"}}))
	assert.Equal(t, presetRateLimitMaxBackoff, rateLimitBackoff(http.Header{"Retry-After": []string{"120"}}))
}

func TestDeleteUser_retriesOnRateLimit(t *testing.T) {
	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount < 3 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.DeleteUser(context.Background(), "user-1"))
	assert.Equal(t, 3, requestCount)
}

func TestDeleteUser_rateLimitExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	err := pc.DeleteUser(context.Background(), "user-1")
	require.Error(t, err)
	assert.True(t, isResponseStatus(err, http.StatusTooManyRequests))
}

func TestCreateUser_conflictResolvesExistingUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == testSCIMPath("/Users"):
			w.WriteHeader(http.StatusConflict)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, testSCIMPath("/Users")):
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(scimUsersResponse{
				TotalResults: 1,
				Resources: []scimUser{{
					ID:       "samlp|redhat|existing@example.com",
					UserName: "existing@example.com",
				}},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	user, err := pc.CreateUser(context.Background(), &structs.User{
		Email:    "existing@example.com",
		UserName: "existing@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "samlp|redhat|existing@example.com", user.ID)
}

func TestDeleteNotFoundIsIdempotent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.DeleteUser(context.Background(), "missing-user"))
}

func TestRemoveUserFromTeam_notFoundReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	err := pc.RemoveUserFromTeam(context.Background(), "group-1", []string{"missing-user"})
	require.Error(t, err)
	assert.True(t, isResponseStatus(err, http.StatusNotFound))
}

func TestAddUserToTeam_conflictReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusConflict)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	err := pc.AddUserToTeam(context.Background(), "group-1", []string{"user-a", "user-b"})
	require.Error(t, err)
	assert.True(t, isResponseStatus(err, http.StatusConflict))
}

func TestRemoveUserFromTeam_batchServiceError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	err := pc.RemoveUserFromTeam(context.Background(), "group-1", []string{"user-a", "user-b"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestEscapeSCIMLiteral(t *testing.T) {
	assert.Equal(t, `user\"name`, escapeSCIMLiteral(`user"name`))
	assert.Equal(t, `path\\to`, escapeSCIMLiteral(`path\to`))
}

func TestNewClient_missingRequiredFields(t *testing.T) {
	poolCfg, hystrixCfg := testHTTPConfigs()

	_, err := NewClient(map[string]interface{}{}, poolCfg, hystrixCfg)
	require.Error(t, err)

	_, err = NewClient(map[string]interface{}{
		"base_url": "https://example.com",
	}, poolCfg, hystrixCfg)
	require.Error(t, err)

	_, err = NewClient(map[string]interface{}{
		"base_url":  "https://example.com",
		"team_slug": "team-1",
	}, poolCfg, hystrixCfg)
	require.Error(t, err)
}

func TestNewClient_success(t *testing.T) {
	poolCfg, hystrixCfg := testHTTPConfigs()

	pc, err := NewClient(map[string]interface{}{
		"base_url":   "https://example.com/",
		"team_slug":  "team-1",
		"scim_token": "token",
	}, poolCfg, hystrixCfg)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", pc.baseURL)
	assert.Equal(t, "team-1", pc.teamSlug)
	assert.Equal(t, "token", pc.scimToken)
}

func TestCreateUser_requiresEmail(t *testing.T) {
	pc := newTestPresetClient(t, "https://example.com")
	_, err := pc.CreateUser(context.Background(), &structs.User{UserName: "no-email"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestCreateTeam_conflictResolvesExistingGroup(t *testing.T) {
	var lookups int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, testSCIMPath("/Groups")):
			lookups++
			w.WriteHeader(http.StatusOK)
			if lookups == 1 {
				_ = json.NewEncoder(w).Encode(scimGroupsResponse{TotalResults: 0})
				return
			}
			_ = json.NewEncoder(w).Encode(scimGroupsResponse{
				TotalResults: 1,
				Resources:    []scimGroup{{ID: "group-conflict", DisplayName: "conflict-team"}},
			})
		case r.Method == http.MethodPost && r.URL.Path == testSCIMPath("/Groups"):
			w.WriteHeader(http.StatusConflict)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	team, err := pc.CreateTeam(context.Background(), &structs.Team{Name: "conflict-team"})
	require.NoError(t, err)
	assert.Equal(t, "group-conflict", team.ID)
}

func TestDeleteNotFoundIsIdempotent_team(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodDelete, r.Method)
		require.True(t, strings.Contains(r.URL.Path, "/Groups/"))
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.DeleteTeamByID(context.Background(), "missing-group"))
}

func TestMembership_emptyInputIsNoOp(t *testing.T) {
	pc := newTestPresetClient(t, "https://example.com")
	require.NoError(t, pc.AddUserToTeam(context.Background(), "group-1", nil))
	require.NoError(t, pc.AddUserToTeam(context.Background(), "group-1", []string{}))
	require.NoError(t, pc.RemoveUserFromTeam(context.Background(), "group-1", nil))
	require.NoError(t, pc.RemoveUserFromTeam(context.Background(), "group-1", []string{}))
}

func TestAddUserToTeam_batchesAt500Limit(t *testing.T) {
	var batchSizes []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		patchReq := decodePatchRequest(t, r)
		batchSizes = append(batchSizes, addBatchMemberCount(patchReq))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	userIDs := makeUserIDs(scimMembershipBatchSize + 1)
	require.NoError(t, pc.AddUserToTeam(context.Background(), "group-1", userIDs))
	assert.Equal(t, []int{scimMembershipBatchSize, 1}, batchSizes)
}

func TestRemoveUserFromTeam_batchesAt500Limit(t *testing.T) {
	var batchSizes []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		patchReq := decodePatchRequest(t, r)
		batchSizes = append(batchSizes, len(patchReq.Operations))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	userIDs := makeUserIDs(scimMembershipBatchSize + 1)
	require.NoError(t, pc.RemoveUserFromTeam(context.Background(), "group-1", userIDs))
	assert.Equal(t, []int{scimMembershipBatchSize, 1}, batchSizes)
}

func TestAddUserToTeam_successSingleBatch(t *testing.T) {
	var patchCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		patchReq := decodePatchRequest(t, r)
		require.Len(t, patchReq.Operations, 1)
		assert.Equal(t, "add", patchReq.Operations[0].Op)
		assert.Equal(t, "members", patchReq.Operations[0].Path)
		assert.Equal(t, 3, addBatchMemberCount(patchReq))
		patchCount++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.AddUserToTeam(context.Background(), "group-1", []string{"u1", "u2", "u3"}))
	assert.Equal(t, 1, patchCount)
}

func TestRemoveUserFromTeam_successSingleBatch(t *testing.T) {
	var patchCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPatch, r.Method)
		patchReq := decodePatchRequest(t, r)
		require.Len(t, patchReq.Operations, 2)
		assert.Equal(t, "remove", patchReq.Operations[0].Op)
		assert.Equal(t, `members[value eq "user-a"]`, patchReq.Operations[0].Path)
		assert.Equal(t, `members[value eq "user-b"]`, patchReq.Operations[1].Path)
		patchCount++
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.RemoveUserFromTeam(context.Background(), "group-1", []string{"user-a", "user-b"}))
	assert.Equal(t, 1, patchCount)
}

func TestRemoveUserFromTeam_escapesReservedCharactersInFilter(t *testing.T) {
	userID := `user"name\path`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		patchReq := decodePatchRequest(t, r)
		require.Len(t, patchReq.Operations, 1)
		assert.Equal(t, fmt.Sprintf(`members[value eq "%s"]`, escapeSCIMLiteral(userID)), patchReq.Operations[0].Path)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	require.NoError(t, pc.RemoveUserFromTeam(context.Background(), "group-1", []string{userID}))
}

func TestAddUserToTeam_continuesAfterBatchFailure(t *testing.T) {
	var batchSizes []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		patchReq := decodePatchRequest(t, r)
		memberCount := addBatchMemberCount(patchReq)
		batchSizes = append(batchSizes, memberCount)
		if memberCount == scimMembershipBatchSize {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	userIDs := makeUserIDs(scimMembershipBatchSize + 1)
	err := pc.AddUserToTeam(context.Background(), "group-1", userIDs)
	require.Error(t, err)
	assert.GreaterOrEqual(t, len(batchSizes), 2)
	assert.Equal(t, scimMembershipBatchSize, batchSizes[0])
	assert.Equal(t, 1, batchSizes[len(batchSizes)-1])
	assert.Contains(t, err.Error(), "batch 1/2")
}

func TestFetchTeamMembersByTeamID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, testSCIMPath("/Groups/group-1"), r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(scimGroup{
			ID: "group-1",
			Members: []scimMember{
				{Value: "user-1", Display: "User One"},
				{Value: "user-2", Display: "User Two"},
			},
		})
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	members, err := pc.FetchTeamMembersByTeamID(context.Background(), "group-1")
	require.NoError(t, err)
	require.Len(t, members, 2)
	assert.Equal(t, "User One", members["user-1"].DisplayName)
	assert.Equal(t, "user-2", members["user-2"].ID)
}

func TestFetchAllTeams_paginates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.True(t, strings.HasPrefix(r.URL.Path, testSCIMPath("/Groups")))

		startIndex := 1
		if raw := r.URL.Query().Get("startIndex"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			require.NoError(t, err)
			startIndex = parsed
		}
		count, err := strconv.Atoi(r.URL.Query().Get("count"))
		require.NoError(t, err)
		require.Equal(t, scimGroupsPageSize, count)

		switch startIndex {
		case 1:
			_ = json.NewEncoder(w).Encode(scimGroupsResponse{
				TotalResults: 2,
				Resources:    []scimGroup{{ID: "g1", DisplayName: "group-one"}},
			})
		case 2:
			_ = json.NewEncoder(w).Encode(scimGroupsResponse{
				TotalResults: 2,
				Resources:    []scimGroup{{ID: "g2", DisplayName: "group-two"}},
			})
		default:
			_ = json.NewEncoder(w).Encode(scimGroupsResponse{TotalResults: 2})
		}
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	teams, err := pc.FetchAllTeams(context.Background())
	require.NoError(t, err)
	require.Len(t, teams, 2)
	assert.Equal(t, "group-one", teams["g1"].Name)
	assert.Equal(t, "group-two", teams["g2"].Name)
}

func TestFetchAllUsers_paginates(t *testing.T) {
	var pageRequests []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, testSCIMPath("/Users"), r.URL.Path)

		startIndex, err := strconv.Atoi(r.URL.Query().Get("startIndex"))
		require.NoError(t, err)
		count, err := strconv.Atoi(r.URL.Query().Get("count"))
		require.NoError(t, err)
		require.Equal(t, scimUsersPageSize, count)
		pageRequests = append(pageRequests, startIndex)

		resources := make([]scimUser, scimUsersPageSize)
		for i := range resources {
			resources[i] = scimUser{
				ID:       fmt.Sprintf("user-%d", startIndex+i),
				UserName: fmt.Sprintf("user-%d@example.com", startIndex+i),
			}
		}

		totalResults := scimUsersPageSize + 1
		if startIndex > 1 {
			resources = resources[:1]
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(scimUsersResponse{
			TotalResults: totalResults,
			Resources:    resources,
		})
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	_, userIDMap, err := pc.FetchAllUsers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []int{1, scimUsersPageSize + 1}, pageRequests)
	assert.Len(t, userIDMap, scimUsersPageSize+1)
}

func TestFindUserByEmail_escapesFilterLiteral(t *testing.T) {
	var filter string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == testSCIMPath("/Users"):
			w.WriteHeader(http.StatusConflict)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, testSCIMPath("/Users")):
			filter = r.URL.Query().Get("filter")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(scimUsersResponse{
				TotalResults: 1,
				Resources: []scimUser{{
					ID:       "samlp|redhat|user@example.com",
					UserName: `user"name@example.com`,
				}},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	t.Cleanup(srv.Close)

	pc := newTestPresetClient(t, srv.URL)
	user, err := pc.CreateUser(context.Background(), &structs.User{
		Email:    `user"name@example.com`,
		UserName: `user"name@example.com`,
	})
	require.NoError(t, err)
	assert.Equal(t, "samlp|redhat|user@example.com", user.ID)
	assert.Contains(t, filter, `user\"name@example.com`)
}
