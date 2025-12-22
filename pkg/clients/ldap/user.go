package ldap

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-ldap/ldap/v3"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

var (
	ErrNoUserFound = errors.New("no LDAP entries found for user")
)

// parseLDAPEntry is a helper method that extracts attribute values from an LDAP entry.
func (l *LDAPConn) parseLDAPEntry(entry *ldap.Entry) map[string]interface{} {
	userData := make(map[string]interface{})
	for _, attr := range l.attributes {
		if len(entry.GetAttributeValues(attr)) > 0 {
			userData[attr] = entry.GetAttributeValue(attr)
		} else {
			userData[attr] = ""
		}
	}
	return userData
}

// executeSearch is a helper method that executes the provided search request.
// It handles connection management, search execution, and result parsing.
func (l *LDAPConn) executeSearch(searchRequest *ldap.SearchRequest) (map[string]interface{}, error) {
	conn := l.getConn()
	if conn == nil {
		return nil, errors.New("LDAP connection is nil")
	}

	// Ensure connection is bound before search (some LDAP servers require this)
	err := conn.UnauthenticatedBind("")
	if err != nil {
		return nil, fmt.Errorf("failed to bind before search: %w", err)
	}

	resp, err := conn.Search(searchRequest)
	if err != nil {
		// Handle LDAP "No Such Object" error (code 32)
		if ldapErr, ok := err.(*ldap.Error); ok {
			if ldapErr.ResultCode == ldap.LDAPResultNoSuchObject {
				return nil, ErrNoUserFound
			}
		}
		return nil, err
	}

	if len(resp.Entries) == 0 {
		return nil, ErrNoUserFound
	}

	return l.parseLDAPEntry(resp.Entries[0]), nil
}

// GetUserLDAPData retrieves user data from LDAP using the userID (username).
// It constructs a search request with a uid filter and performs a subtree search in baseDN.
func (l *LDAPConn) GetUserLDAPData(ctx context.Context, userID string) (map[string]interface{}, error) {
	log := logger.Logger(ctx).WithField("userID", userID)
	log.Debug("fetching user LDAP data")

	filter := fmt.Sprintf("(%s)", l.userSearchFilter)

	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf(l.userDN, ldap.EscapeFilter(userID)),
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		l.attributes,
		nil,
	)

	userData, err := l.executeSearch(searchRequest)
	if err != nil {
		if err == ErrNoUserFound {
			log.Warn("no LDAP entries found for user")
		} else {
			log.WithError(err).Error("failed to search LDAP for user data")
		}
		return nil, err
	}

	log.Debug("fetched user LDAP data")
	return userData, nil
}

// GetUserLDAPDataByEmail retrieves user data from LDAP using the email address.
// It constructs a search request with a mail filter and performs a subtree search in baseDN.
func (l *LDAPConn) GetUserLDAPDataByEmail(ctx context.Context, email string) (map[string]interface{}, error) {
	log := logger.Logger(ctx).WithField("email", email)
	log.Debug("fetching user LDAP data by email")

	// Construct search filter: (&userSearchFilter (mail=email))
	mailFilter := fmt.Sprintf("(mail=%s)", ldap.EscapeFilter(email))
	filter := fmt.Sprintf("(&%s%s)", l.userSearchFilter, mailFilter)

	searchRequest := ldap.NewSearchRequest(
		l.BaseUserDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		l.attributes,
		nil,
	)

	userData, err := l.executeSearch(searchRequest)
	if err != nil {
		if err == ErrNoUserFound {
			log.Warn("no LDAP entries found for email")
		} else {
			log.WithError(err).Error("failed to search LDAP for user data by email")
		}
		return nil, err
	}

	log.Debug("fetched user LDAP data by email")
	return userData, nil
}
