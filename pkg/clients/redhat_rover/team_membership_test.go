package redhatrover

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordedRequest struct {
	method string
	url    string
	body   []byte
}

type fakeDoer struct {
	status int
	body   string
	err    error
	reqs   []recordedRequest
}

func (f *fakeDoer) Do(r *http.Request) (*http.Response, error) {
	var buf []byte
	if r.Body != nil {
		buf, _ = io.ReadAll(r.Body)
	}
	f.reqs = append(f.reqs, recordedRequest{method: r.Method, url: r.URL.String(), body: buf})
	if f.err != nil {
		return nil, f.err
	}
	status := f.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(f.body)),
		Header:     make(http.Header),
		Request:    r,
	}, nil
}

func testGroupBody() string {
	return `{
		"name": "aif-group",
		"members": [
			{"id": "alice", "type": "user"},
			{"id": "usernaut-prod", "type": "serviceaccount"},
			{"id": "nested", "type": "group"}
		]
	}`
}

func newTestClient(doer *fakeDoer) *RoverClient {
	return &RoverClient{client: doer, url: "https://rover.test"}
}

func TestFetchTeamMembersFiltersServiceAccounts(t *testing.T) {
	doer := &fakeDoer{body: testGroupBody()}
	client := newTestClient(doer)

	without, err := client.FetchTeamMembersByTeamID(context.Background(), "aif-group")
	require.NoError(t, err)
	require.Len(t, without, 1)
	assert.Equal(t, "alice", without["alice"].ID)
	assert.Equal(t, MemberTypeServiceAccount, client.memberType("aif-group", "usernaut-prod"))
	assert.Equal(t, "group", client.memberType("aif-group", "nested"))

	withSA, err := client.FetchTeamMembersByTeamID(config.WithAllowServiceAccounts(context.Background()), "aif-group")
	require.NoError(t, err)
	require.Len(t, withSA, 2)
	assert.Equal(t, "alice", withSA["alice"].ID)
	assert.Equal(t, "usernaut-prod", withSA["usernaut-prod"].ID)
	_, hasGroup := withSA["nested"]
	assert.False(t, hasGroup)
}

func TestAddAndRemoveUseRememberedMemberType(t *testing.T) {
	doer := &fakeDoer{body: testGroupBody()}
	client := newTestClient(doer)
	ctx := config.WithAllowServiceAccounts(context.Background())

	_, err := client.FetchTeamMembersByTeamID(ctx, "aif-group")
	require.NoError(t, err)

	require.NoError(t, client.AddUserToTeam(ctx, "aif-group", []string{"usernaut-prod", "alice"}))
	require.NoError(t, client.RemoveUserFromTeam(ctx, "aif-group", []string{"usernaut-prod"}))

	var addReq, delReq MemberModRequest
	require.NoError(t, json.Unmarshal(doer.reqs[1].body, &addReq))
	require.NoError(t, json.Unmarshal(doer.reqs[2].body, &delReq))

	assert.Equal(t, []Member{
		{ID: "usernaut-prod", Type: MemberTypeServiceAccount},
		{ID: "alice", Type: MemberTypeUser},
	}, addReq.Additions)
	assert.Empty(t, addReq.Deletions)
	assert.Equal(t, []Member{
		{ID: "usernaut-prod", Type: MemberTypeServiceAccount},
	}, delReq.Deletions)
	assert.Empty(t, delReq.Additions)
	assert.Contains(t, doer.reqs[1].url, "/v1/groups/aif-group/membersMod")
}

func TestAddUnknownMemberDefaultsToUser(t *testing.T) {
	doer := &fakeDoer{body: `{"members":[]}`}
	client := newTestClient(doer)

	require.NoError(t, client.AddUserToTeam(context.Background(), "aif-group", []string{"usernaut-prod"}))

	var addReq MemberModRequest
	require.NoError(t, json.Unmarshal(doer.reqs[0].body, &addReq))
	assert.Equal(t, []Member{{ID: "usernaut-prod", Type: MemberTypeUser}}, addReq.Additions)
}

func TestFetchTeamMembersByTeamIDErrors(t *testing.T) {
	ctx := context.Background()

	errClient := newTestClient(&fakeDoer{err: assert.AnError})
	_, err := errClient.FetchTeamMembersByTeamID(ctx, "aif-group")
	require.Error(t, err)

	statusClient := newTestClient(&fakeDoer{status: http.StatusForbidden, body: "nope"})
	_, err = statusClient.FetchTeamMembersByTeamID(ctx, "aif-group")
	require.EqualError(t, err, "failed to fetch rover group members with response code: Forbidden")

	jsonClient := newTestClient(&fakeDoer{body: "not-json"})
	_, err = jsonClient.FetchTeamMembersByTeamID(ctx, "aif-group")
	require.ErrorContains(t, err, "failed to decode rover group response")
}

func TestModifyErrors(t *testing.T) {
	ctx := context.Background()
	client := newTestClient(&fakeDoer{})
	require.EqualError(t, client.modify(ctx, "span", "nope", "aif-group", []string{"alice"}), "invalid action:nope")

	errClient := newTestClient(&fakeDoer{err: assert.AnError})
	require.Error(t, errClient.AddUserToTeam(ctx, "aif-group", []string{"alice"}))

	statusClient := newTestClient(&fakeDoer{status: http.StatusBadRequest, body: "bad"})
	err := statusClient.RemoveUserFromTeam(ctx, "aif-group", []string{"alice"})
	require.EqualError(t, err, "failed to remove users in rover group with response code: Bad Request")
}

func TestRememberMemberTypesReplacesTeamMap(t *testing.T) {
	client := &RoverClient{}
	client.rememberMemberTypes("aif-group", []Member{{ID: "svc", Type: MemberTypeServiceAccount}})
	assert.Equal(t, MemberTypeServiceAccount, client.memberType("aif-group", "svc"))

	client.rememberMemberTypes("aif-group", []Member{{ID: "alice", Type: MemberTypeUser}})
	assert.Equal(t, MemberTypeUser, client.memberType("aif-group", "alice"))
	assert.Equal(t, MemberTypeUser, client.memberType("aif-group", "svc"))
}
