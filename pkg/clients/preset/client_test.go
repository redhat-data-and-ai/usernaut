package preset

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/heimdall/v7"
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

func TestFetchTeamMembersByTeamID_singleResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, testSCIMPath("/Groups/group-1"), r.URL.Path)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(scimGroup{
			ID:           "group-1",
			TotalResults: 2,
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
