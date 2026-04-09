package ldap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

const bulkLDAPBatchSize = 500

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
func (l *LDAPConn) executeSearch(ctx context.Context,
	searchRequest *ldap.SearchRequest) (map[string]interface{}, error) {
	log := logger.Logger(ctx).WithField("searchRequest", searchRequest)
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
				log.WithError(err).Debug("LDAP Result Code 32: No Such Object")
				return nil, ErrNoUserFound
			}
		}
		return nil, err
	}

	if len(resp.Entries) == 0 {
		log.Warn("no LDAP entries found")
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

	userData, err := l.executeSearch(ctx, searchRequest)
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

// GetBulkUserLDAPData retrieves LDAP data for multiple users in batched OR queries,
// returning a map keyed by uid. Users not found in LDAP are silently omitted from
// the result. Batches are capped at bulkLDAPBatchSize to stay within server size limits.
func (l *LDAPConn) GetBulkUserLDAPData(
	ctx context.Context,
	userIDs []string,
) (map[string]map[string]interface{}, error) {
	log := logger.Logger(ctx)
	result := make(map[string]map[string]interface{}, len(userIDs))

	if len(userIDs) == 0 {
		return result, nil
	}

	totalBatches := (len(userIDs) + bulkLDAPBatchSize - 1) / bulkLDAPBatchSize
	log.WithField("total_users", len(userIDs)).WithField("batch_size", bulkLDAPBatchSize).
		WithField("total_batches", totalBatches).Info("starting bulk LDAP fetch")

	for batchStart := 0; batchStart < len(userIDs); batchStart += bulkLDAPBatchSize {
		batchEnd := batchStart + bulkLDAPBatchSize
		if batchEnd > len(userIDs) {
			batchEnd = len(userIDs)
		}
		batch := userIDs[batchStart:batchEnd]
		batchNum := (batchStart / bulkLDAPBatchSize) + 1

		log.WithField("batch", fmt.Sprintf("%d/%d", batchNum, totalBatches)).
			WithField("batch_users", len(batch)).Info("processing LDAP batch")

		var uidFilters strings.Builder
		for _, uid := range batch {
			fmt.Fprintf(&uidFilters, "(uid=%s)", ldap.EscapeFilter(uid))
		}
		filter := fmt.Sprintf("(&(%s)(|%s))", l.userSearchFilter, uidFilters.String())

		searchRequest := ldap.NewSearchRequest(
			l.baseUserDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			filter,
			l.attributes,
			nil,
		)

		conn := l.getConn()
		if conn == nil {
			return result, errors.New("LDAP connection is nil")
		}

		if err := conn.UnauthenticatedBind(""); err != nil {
			return result, fmt.Errorf("failed to bind before bulk search: %w", err)
		}

		resp, err := conn.Search(searchRequest)
		if err != nil {
			log.WithError(err).WithField("batch_start", batchStart).Error("bulk LDAP search failed")
			return result, err
		}

		log.WithField("batch_start", batchStart).WithField("entries", len(resp.Entries)).
			Info("bulk LDAP batch returned")

		for _, entry := range resp.Entries {
			uid := entry.GetAttributeValue("uid")
			if uid == "" {
				continue
			}
			result[uid] = l.parseLDAPEntry(entry)
		}
	}

	return result, nil
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
		l.baseUserDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		l.attributes,
		nil,
	)

	userData, err := l.executeSearch(ctx, searchRequest)
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
