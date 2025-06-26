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

package redhatrover

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemberGetters(t *testing.T) {
	m := Member{Type: MemberTypeUser, ID: "user-123"}
	assert.Equal(t, MemberTypeUser, m.GetType())
	assert.Equal(t, "user-123", m.GetID())
}

func TestRoverGroupGetters(t *testing.T) {
	displayName := "Display Name"
	memberOf := "parent-group"
	namespace := "test-namespace"
	member := Member{Type: MemberTypeServiceAccount, ID: "svc-1"}
	group := RoverGroup{
		Name:                  "group1",
		Description:           "desc",
		MemberApprovalType:    MemberApprovalTypeSelfService,
		Owners:                []Member{member},
		ContactList:           "contact@example.com",
		DisplayName:           &displayName,
		Notes:                 "some notes",
		RoverGroupMemberQuery: nil,
		RoverGroupInclusions:  []Member{member},
		RoverGroupExclusions:  []Member{},
		Members:               []Member{member},
		MemberOf:              &memberOf,
		Namespace:             &namespace,
	}
	assert.Equal(t, "group1", group.GetName())
	assert.Equal(t, "desc", group.GetDescription())
	assert.Equal(t, MemberApprovalTypeSelfService, group.GetMemberApprovalType())
	assert.Equal(t, []Member{member}, group.GetOwners())
	assert.Equal(t, "contact@example.com", group.GetContactList())
	assert.Equal(t, &displayName, group.GetDisplayName())
	assert.Equal(t, "some notes", group.GetNotes())
	assert.Nil(t, group.GetRoverGroupMemberQuery())
	assert.Equal(t, []Member{member}, group.GetRoverGroupInclusions())
	assert.Equal(t, []Member{}, group.GetRoverGroupExclusions())
	assert.Equal(t, []Member{member}, group.GetMembers())
	assert.Equal(t, &memberOf, group.GetMemberOf())
	assert.Equal(t, &namespace, group.GetNamespace())
}

func TestMemberModRequestGetters(t *testing.T) {
	add := Member{Type: MemberTypeUser, ID: "add-1"}
	del := Member{Type: MemberTypeServiceAccount, ID: "del-1"}
	modReq := MemberModRequest{
		Additions: []Member{add},
		Deletions: []Member{del},
	}
	assert.Equal(t, []Member{add}, modReq.GetAdditions())
	assert.Equal(t, []Member{del}, modReq.GetDeletions())
}
