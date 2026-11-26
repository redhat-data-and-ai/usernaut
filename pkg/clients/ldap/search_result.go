package ldap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

// ErrLDAPDiagnosticMessage is returned when LDAP server is in reinit state, it
// may also be returned when the LDAP server is in a non-healthy state (busy, unavilable, etc.)
var ErrLDAPDiagnosticMessage = errors.New("LDAP returned diagnostic message")

// validateSearchResultResponse returns ErrLDAPDiagnosticMessage when the
// search response includes a diagnostic message.
func validateSearchResultResponse(ctx context.Context, resp *ldap.SearchResult) error {
	log := logger.Logger(ctx)

	code := resp.ResultCode
	diagnostic := strings.TrimSpace(resp.DiagnosticMessage)

	// This condition is true when the LDAP server is in reinit state.
	if diagnostic != "" && len(resp.Entries) == 0 {
		log.WithField("ldap_diagnostic_message", diagnostic).
			WithField("ldap_result_code", code).
			Error("aborting: LDAP returned diagnostic message with zero entries")
		return fmt.Errorf("%w: resultCode=%d diagnostic=%q",
			ErrLDAPDiagnosticMessage, code, diagnostic)
	}

	// This condition is true when the LDAP server is in a non-healthy state (busy, unavilable, etc.)
	if diagnostic != "" {
		log.WithField("ldap_diagnostic_message", diagnostic).
			WithField("ldap_result_code", code).
			Error("aborting: LDAP returned diagnostic message with non-zero entries")
		return fmt.Errorf("%w: resultCode=%d diagnostic=%q",
			ErrLDAPDiagnosticMessage, code, diagnostic)
	}
	return nil
}
