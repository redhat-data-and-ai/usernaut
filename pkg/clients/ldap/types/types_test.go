package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUIDFromManagerDN(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "pbhattac", uidFromManagerDN("uid=pbhattac,ou=users,dc=redhat,dc=com"))
	assert.Equal(t, "pbhattac", uidFromManagerDN("pbhattac"))
	assert.Equal(t, "", uidFromManagerDN(""))
}

func TestQueryMembers_Managers(t *testing.T) {
	t.Parallel()
	members := QueryMembers{
		{UID: "alice", Manager: "uid=pbhattac,ou=users,dc=redhat,dc=com"},
		{UID: "bob", Manager: "uid=otherboss,ou=users,dc=redhat,dc=com"},
		{UID: "carol", Manager: "uid=pbhattac,ou=users,dc=redhat,dc=com"},
		{UID: "dave", Manager: ""},
	}
	got := members.Managers()
	assert.ElementsMatch(t, []string{"pbhattac", "otherboss"}, got.UIDs())
}

func TestQueryMembers_Dedupe(t *testing.T) {
	t.Parallel()
	members := QueryMembers{
		{UID: "alice", Manager: "m1"},
		{UID: "bob"},
		{UID: "alice", Manager: "m2"},
		{UID: ""},
	}
	got := members.Dedupe()
	assert.Equal(t, QueryMembers{{UID: "alice", Manager: "m1"}, {UID: "bob"}}, got)
}
