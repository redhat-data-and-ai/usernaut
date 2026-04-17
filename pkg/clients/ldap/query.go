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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
	log.WithField("query", query).Info("building LDAP query from spec")

	if query == nil {
		return "", errors.New("ldap query is nil")
	}
	if len(query.Filters) == 0 {
		return "", errors.New("filters are empty")
	}
	filters, err := buildFiltersFromSpec(query.Filters, l.baseUserDN)
	if err != nil {
		return "", err
	}

	op := strings.ToLower(strings.TrimSpace(query.Operator))
	switch op {
	case "and":
		return "(&" + strings.Join(filters, "") + ")", nil
	case "or":
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

	if baseUserDN == "" {
		return "", errors.New("base user DN is empty")
	}

	key := strings.TrimSpace(filter.Key)
	if key == "" {
		return "", errors.New("filter key is empty")
	}

	value := ldap.EscapeFilter(strings.TrimSpace(filter.Value))
	if value == "" {
		return "", errors.New("filter value is empty")
	}

	if op == "contains" {
		value = "*" + value + "*"
	}

	if strings.EqualFold(key, "manager") {
		value = "uid=" + value + "," + baseUserDN
	}

	switch op {
	case "equals":
		return "(" + key + "=" + value + ")", nil
	case "not":
		return "(!(" + key + "=" + value + "))", nil
	case "contains":
		return "(" + key + "=" + value + ")", nil
	default:
		return "", fmt.Errorf("unsupported filter operator %q", op)
	}
}
