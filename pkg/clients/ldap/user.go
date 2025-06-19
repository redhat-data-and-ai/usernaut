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

func (l *LDAPConn) GetUserLDAPData(ctx context.Context, userID string) (map[string]interface{}, error) {
	log := logger.Logger(ctx).WithField("userID", userID)
	log.Info("fetching user LDAP data")

	searchRequest := ldap.NewSearchRequest(
		fmt.Sprintf(l.userDN, ldap.EscapeFilter(userID)),
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		l.userSearchFilter,
		l.attributes,
		nil,
	)

	conn := l.getConn()
	if conn == nil {
		log.Error("LDAP connection is nil, cannot perform search")
		return nil, errors.New("LDAP connection is nil")
	}

	resp, err := conn.Search(searchRequest)
	if err != nil {
		log.WithError(err).Error("failed to search LDAP for user data")
		return nil, err
	}
	if len(resp.Entries) == 0 {
		log.Warn("no LDAP entries found for user")
		return nil, ErrNoUserFound
	}
	userData := make(map[string]interface{})
	for _, attr := range l.attributes {
		if len(resp.Entries[0].GetAttributeValues(attr)) > 0 {
			userData[attr] = resp.Entries[0].GetAttributeValue(attr)
		} else {
			userData[attr] = ""
		}
	}

	log.Info("fetched user LDAP data")
	return userData, nil
}
