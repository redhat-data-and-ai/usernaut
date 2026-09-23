package ldap

import (
	"errors"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchResultDonePacket(resultCode int64, diagnostic string) *ber.Packet {
	done := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ldap.ApplicationSearchResultDone, nil, "Search Result Done")
	done.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultCode, "resultCode"))
	done.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matchedDN"))
	done.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, diagnostic, "diagnosticMessage"))

	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Message")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, 1, "messageID"))
	packet.AppendChild(done)
	return packet
}

func TestParseLDAPResult_SuccessKeepsDiagnostic(t *testing.T) {
	packet := searchResultDonePacket(0, "IPA: server is reinitializing")

	require.Nil(t, ldap.GetLDAPError(packet), "go-ldap treats resultCode 0 as success and returns nil")

	code, _, diagnostic := ldap.ParseLDAPResult(packet)
	assert.Equal(t, uint16(ldap.LDAPResultSuccess), code)
	assert.Equal(t, "IPA: server is reinitializing", diagnostic)
}

func TestParseLDAPResult_NonZeroKeepsDiagnostic(t *testing.T) {
	packet := searchResultDonePacket(int64(ldap.LDAPResultBusy), "server busy")

	err := ldap.GetLDAPError(packet)
	require.Error(t, err)
	var ldapErr *ldap.Error
	require.ErrorAs(t, err, &ldapErr)
	assert.Equal(t, uint16(ldap.LDAPResultBusy), ldapErr.ResultCode)
	assert.Equal(t, "server busy", ldapErr.Err.Error())

	code, _, diagnostic := ldap.ParseLDAPResult(packet)
	assert.Equal(t, uint16(ldap.LDAPResultBusy), code)
	assert.Equal(t, "server busy", diagnostic)
}

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
