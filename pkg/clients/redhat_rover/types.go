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

import "github.com/redhat-data-and-ai/usernaut/pkg/common/constants"

const (
	MemberApprovalTypeSelfService = "self-service"
	MemberTypeUser                = "user"
	MemberTypeServiceAccount      = "serviceaccount"
	defaultContactEmail           = "devnull@redhat.com"
)

var (
	headers = map[string]string{constants.ContentTypeHeaderKey: "application/json"}
)

type Member struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type RoverGroup struct {
	Name                  string   `json:"name"`
	Description           string   `json:"description"`
	MemberApprovalType    string   `json:"memberApprovalType"`
	Owners                []Member `json:"owners"`
	ContactList           string   `json:"contactList"`
	DisplayName           *string  `json:"displayName"`
	Notes                 string   `json:"notes"`
	RoverGroupMemberQuery *string  `json:"roverGroupMemberQuery"`
	RoverGroupInclusions  []Member `json:"roverGroupInclusions"`
	RoverGroupExclusions  []Member `json:"roverGroupExclusions"`
	Members               []Member `json:"members"`
	MemberOf              *string  `json:"memberOf"`
	Namespace             *string  `json:"namespace"`
}

// Member getters
func (m *Member) GetType() string {
	return m.Type
}

func (m *Member) GetID() string {
	return m.ID
}

// RoverGroup getters
func (r *RoverGroup) GetName() string {
	return r.Name
}

func (r *RoverGroup) GetDescription() string {
	return r.Description
}

func (r *RoverGroup) GetMemberApprovalType() string {
	return r.MemberApprovalType
}

func (r *RoverGroup) GetOwners() []Member {
	return r.Owners
}

func (r *RoverGroup) GetContactList() string {
	return r.ContactList
}

func (r *RoverGroup) GetDisplayName() *string {
	return r.DisplayName
}

func (r *RoverGroup) GetNotes() string {
	return r.Notes
}

func (r *RoverGroup) GetRoverGroupMemberQuery() *string {
	return r.RoverGroupMemberQuery
}

func (r *RoverGroup) GetRoverGroupInclusions() []Member {
	return r.RoverGroupInclusions
}

func (r *RoverGroup) GetRoverGroupExclusions() []Member {
	return r.RoverGroupExclusions
}

func (r *RoverGroup) GetMembers() []Member {
	return r.Members
}

func (r *RoverGroup) GetMemberOf() *string {
	return r.MemberOf
}

func (r *RoverGroup) GetNamespace() *string {
	return r.Namespace
}

type MemberModRequest struct {
	Additions []Member `json:"additions"`
	Deletions []Member `json:"deletions"`
}

func (m *MemberModRequest) GetAdditions() []Member {
	return m.Additions
}

func (m *MemberModRequest) GetDeletions() []Member {
	return m.Deletions
}
