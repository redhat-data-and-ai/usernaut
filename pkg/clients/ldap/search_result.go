package ldap

import (
	"errors"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// ErrLDAPUntrustworthyResult is returned when Search completed without a
// library error (often resultCode 0) but the response cannot be treated as a
// complete user list. IPA may return success with an empty entry list and a
// diagnostic message while the directory is not ready (for example during reinit).
var ErrLDAPUntrustworthyResult = errors.New("LDAP search result is untrustworthy")

// searchResultMeta extracts resultCode, diagnosticMessage and entry count from
// a Search reply. It reads from the SearchResult struct first, then falls back
// to the ldap.Error if present (e.g. when the server returned a non-zero code).
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
