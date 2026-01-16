package ldap

import (
	"context"
	"errors"

	"github.com/go-ldap/ldap/v3"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

func (l *LDAPConn) GetQueryMembers(ctx context.Context, query string) ([]string, error) {
	log := logger.Logger(ctx).WithField("query", query)
	log.Info("fetching query members")

	// Empty query means "no query-based members".
	if query == "" {
		log.Info("empty query provided; returning no query members")
		return []string{}, nil
	}

	searchRequest := ldap.NewSearchRequest(
		l.baseUserDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query,
		[]string{"uid"},
		nil,
	)

	conn := l.getConn()
	if conn == nil {
		log.Error("LDAP connection is nil, cannot perform search")
		return nil, errors.New("LDAP connection is nil")
	}
	resp, err := conn.Search(searchRequest)
	if err != nil {
		log.WithError(err).Error("failed to search LDAP for query members")
		return nil, err
	}
	log.WithField("entries", len(resp.Entries)).Info("LDAP search results")
	if len(resp.Entries) == 0 {
		log.Info("no LDAP entries found for query; returning empty member list")
		return []string{}, nil
	}
	queryMembers := make([]string, 0, len(resp.Entries))
	for _, entry := range resp.Entries {
		uid := entry.GetAttributeValue("uid")
		if uid == "" {
			// Fallback: parse uid from DN if attribute is not returned for some reason.
			if dn, parseErr := ldap.ParseDN(entry.DN); parseErr == nil {
				uid = parseUIDFromDN(dn)
			}
		}
		if uid != "" {
			queryMembers = append(queryMembers, uid)
		}
	}
	log.Info("fetched query members")
	return queryMembers, nil
}

func parseUIDFromDN(dn *ldap.DN) string {
	for _, rdn := range dn.RDNs {
		for _, atv := range rdn.Attributes {
			if atv.Type == "uid" && atv.Value != "" {
				return atv.Value
			}
		}
	}
	return ""
}
