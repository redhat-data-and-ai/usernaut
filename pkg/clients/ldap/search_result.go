package ldap

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/sirupsen/logrus"
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

// checkDiagnostic logs a warning when the LDAP response contains a diagnostic message
// and returns an error if the response also has zero entries.
// Callers should enrich the logger with context-specific fields (e.g. batch number)
// before passing it.
func checkDiagnostic(log *logrus.Entry, resp *ldap.SearchResult) error {
	_, diagnostic, _ := searchResultMeta(resp, nil)
	if diagnostic == "" {
		return nil
	}
	log.WithField("ldap_diagnostic_message", diagnostic).
		WithField("ldap_result_code", resp.ResultCode).
		WithField("ldap_entries", len(resp.Entries)).
		Warn("LDAP search returned diagnostic message")
	if len(resp.Entries) == 0 {
		log.WithField("ldap_diagnostic_message", diagnostic).
			WithField("ldap_result_code", resp.ResultCode).
			Error("aborting: LDAP returned diagnostic message with zero entries")
		return fmt.Errorf("%w: resultCode=%d diagnostic=%q",
			ErrLDAPDiagnosticMessage, resp.ResultCode, diagnostic)
	}
	return nil
}
