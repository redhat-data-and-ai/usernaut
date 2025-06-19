package ldap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitLdap_Error(t *testing.T) {
	LDAPConfig := LDAP{
		Server:           "ldap://ldap.com:389",
		BaseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		UserDN:           "uid=%s,ou=users,dc=example,dc=com",
		UserSearchFilter: "(objectClass=uid)",
		Attributes:       []string{"mail"},
	}

	_, err := InitLdap(LDAPConfig)
	assert.Error(t, err, "Expected error due to missing LDAP server connection")
}

func TestInitLdap_Success(t *testing.T) {
	LDAPConfig := LDAP{
		// using a valid LDAP server for testing, reference: https://github.com/go-ldap/ldap/blob/master/v3/ldap_test.go#L13
		Server:           "ldap://ldap.itd.umich.edu:389",
		BaseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		UserDN:           "uid=%s,ou=users,dc=example,dc=com",
		UserSearchFilter: "(objectClass=uid)",
		Attributes:       []string{"mail"},
	}

	client, err := InitLdap(LDAPConfig)
	assert.NoError(t, err, "Expected successful LDAP client initialization")
	assert.NotNil(t, client, "Expected non-nil LDAP client")
}
