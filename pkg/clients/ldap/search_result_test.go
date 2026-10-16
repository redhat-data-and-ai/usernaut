package ldap

import (
	"errors"
	"testing"

	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
)

func TestSearchResultMeta_FromSearchResultOnSuccess(t *testing.T) {
	resp := &ldap.SearchResult{
		ResultCode:        ldap.LDAPResultSuccess,
		DiagnosticMessage: "  IPA reinit  ",
		Entries:           []*ldap.Entry{},
	}
	code, diagnostic, n := searchResultMeta(resp, nil)
	assert.Equal(t, uint16(ldap.LDAPResultSuccess), code)
	assert.Equal(t, "IPA reinit", diagnostic)
	assert.Equal(t, 0, n)
}

func TestSearchResultMeta_FromLDAPError(t *testing.T) {
	err := ldap.NewError(ldap.LDAPResultUnavailable, errors.New("not ready"))
	code, diagnostic, n := searchResultMeta(nil, err)
	assert.Equal(t, uint16(ldap.LDAPResultUnavailable), code)
	assert.Equal(t, "not ready", diagnostic)
	assert.Equal(t, 0, n)
}
