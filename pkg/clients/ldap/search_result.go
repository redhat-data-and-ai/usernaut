package ldap

import (
	"errors"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// ErrLDAPDiagnosticMessage is returned when LDAP search succeeded but the response
// contains a diagnostic message, indicating the server may not be in a healthy state.
var ErrLDAPDiagnosticMessage = errors.New("LDAP returned diagnostic message")

// searchResultMeta extracts resultCode, diagnosticMessage and entry count from a Search reply.
// Falls back to ldap.Error fields when the response is nil or missing values.
func searchResultMeta(resp *ldap.SearchResult, err error) (code uint16, diagnostic string, n int) {
	if resp != nil {
		code = resp.ResultCode
		diagnostic = strings.TrimSpace(resp.DiagnosticMessage)
		n = len(resp.Entries)
	}
	if err != nil {
		var ldapErr *ldap.Error
		if errors.As(err, &ldapErr) {
			if code == 0 {
				code = ldapErr.ResultCode
			}
			if diagnostic == "" && ldapErr.Err != nil {
				diagnostic = strings.TrimSpace(ldapErr.Err.Error())
			}
		}
	}
	return code, diagnostic, n
}
