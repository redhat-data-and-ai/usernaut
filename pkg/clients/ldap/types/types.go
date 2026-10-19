package types

import (
	"strings"
)

// QueryMember is one LDAP search result used for group membership resolution.
type QueryMember struct {
	// UID of the query member
	UID string
	// raw manager DN (or empty)
	Manager string
}

// QueryMembers is a list of QueryMember.
type QueryMembers []QueryMember

// UIDs returns the UID of each query member.
func (q QueryMembers) UIDs() []string {
	uids := make([]string, 0, len(q))
	for _, member := range q {
		if member.UID == "" {
			continue
		}
		uids = append(uids, member.UID)
	}
	return uids
}

// Dedupe returns members with unique UIDs, preserving first occurrence.
func (q QueryMembers) Dedupe() QueryMembers {
	seen := make(map[string]struct{}, len(q))
	deduped := make(QueryMembers, 0, len(q))
	for _, member := range q {
		if member.UID == "" {
			continue
		}
		if _, ok := seen[member.UID]; ok {
			continue
		}
		seen[member.UID] = struct{}{}
		deduped = append(deduped, member)
	}
	return deduped
}

// Managers returns unique manager UIDs parsed from each member's Manager DN.
func (q QueryMembers) Managers() QueryMembers {
	managers := make(QueryMembers, 0, len(q))
	seen := make(map[string]struct{}, len(q))
	for _, member := range q {
		mgrUID := uidFromManagerDN(member.Manager)
		if mgrUID == "" {
			continue
		}
		if _, ok := seen[mgrUID]; ok {
			continue
		}
		seen[mgrUID] = struct{}{}
		managers = append(managers, QueryMember{UID: mgrUID})
	}
	return managers
}

func uidFromManagerDN(manager string) string {
	manager = strings.TrimSpace(manager)
	if manager == "" {
		return ""
	}
	lower := strings.ToLower(manager)
	idx := strings.Index(lower, "uid=")
	if idx < 0 {
		if !strings.Contains(manager, "=") {
			return manager
		}
		return ""
	}
	rest := manager[idx+4:]
	if comma := strings.Index(rest, ","); comma >= 0 {
		rest = rest[:comma]
	}
	return strings.TrimSpace(rest)
}
