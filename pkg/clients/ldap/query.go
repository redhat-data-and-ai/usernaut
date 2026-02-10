package ldap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
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

func (l *LDAPConn) BuildLDAPQueryFromSpec(ctx context.Context, query *v1alpha1.LDAPQuery) (string, error) {
	log := logger.Logger(ctx).WithField("build_ldap_query", "spec")
	filter, err := buildLDAPQueryFromSpec(query, l.baseUserDN)
	if err != nil {
		log.WithError(err).Error("failed to build ldap query from spec")
		return "", err
	}

	return filter, nil
}

func buildLDAPQueryFromSpec(query *v1alpha1.LDAPQuery, baseUserDN string) (string, error) {
	if query == nil {
		return "", errors.New("ldap query is nil")
	}

	op := strings.ToLower(strings.TrimSpace(query.Operator))
	switch op {
	case "and":
		if len(query.Filters) == 0 {
			return "", errors.New("and operator requires filters")
		}
		filters, err := buildFiltersFromSpec(query.Filters, baseUserDN)
		if err != nil {
			return "", err
		}
		return "(&" + strings.Join(filters, "") + ")", nil
	case "or":
		if len(query.Filters) == 0 {
			return "", errors.New("or operator requires filters")
		}
		filters, err := buildFiltersFromSpec(query.Filters, baseUserDN)
		if err != nil {
			return "", err
		}
		return "(|" + strings.Join(filters, "") + ")", nil
	default:
		return "", fmt.Errorf("unsupported operator %q", query.Operator)
	}
}

func buildFiltersFromSpec(filters []v1alpha1.LDAPFilter, baseUserDN string) ([]string, error) {
	results := make([]string, 0, len(filters))
	for _, filter := range filters {
		result, err := buildFilterFromSpec(filter, baseUserDN)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func buildFilterFromSpec(filter v1alpha1.LDAPFilter, baseUserDN string) (string, error) {
	op := strings.ToLower(strings.TrimSpace(filter.Criteria))
	switch op {
	case "equals":
		return buildEqualsFilter(filter.Key, filter.Value, baseUserDN)
	case "not":
		return buildNotFilter(filter.Key, filter.Value, baseUserDN)
	case "contains":
		return buildContainsFilter(filter.Key, filter.Value, baseUserDN)
	default:
		return "", fmt.Errorf("unsupported filter operator %q", filter.Criteria)
	}
}

func buildNotFilter(key, value, baseUserDN string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("not operator requires key")
	}
	if strings.TrimSpace(value) == "" {
		return "", errors.New("not operator requires value")
	}
	return "(!(" + key + "=" + value + "))", nil
}

func buildContainsFilter(key, value, baseUserDN string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("contains operator requires key")
	}
	if strings.TrimSpace(value) == "" {
		return "", errors.New("contains operator requires value")
	}
	return "(" + key + "=*" + value + "*)", nil
}

func buildEqualsFilter(attr, value, baseUserDN string) (string, error) {
	if strings.TrimSpace(attr) == "" {
		return "", errors.New("equals operator requires attr")
	}
	if strings.TrimSpace(value) == "" {
		return "", errors.New("equals operator requires value")
	}

	value = strings.TrimSpace(value)
	if attr == "manager" && !strings.Contains(value, ",") && baseUserDN != "" {
		value = "uid=" + value + "," + baseUserDN
	}

	return "(" + attr + "=" + value + ")", nil
}
