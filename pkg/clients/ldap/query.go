package ldap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients/ldap/types"
	"github.com/redhat-data-and-ai/usernaut/pkg/logger"
)

var defaultAttributes = []string{"uid", "manager"}

// GetQueryMembers searches LDAP with the given filter and returns QueryMember DTOs
// (uid + manager attributes).
func (l *LDAPConn) GetQueryMembers(ctx context.Context, query string) (types.QueryMembers, error) {
	// Do not call LDAP if the request context is already cancelled or its deadline has passed
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	log := logger.Logger(ctx).WithField("query", query)
	log.Info("fetching query members")

	// Empty query means "no query-based members".
	if query == "" {
		log.Info("empty query provided; returning no query members")
		return types.QueryMembers{}, nil
	}

	searchRequest := ldap.NewSearchRequest(
		l.baseUserDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		query,
		defaultAttributes,
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
		return types.QueryMembers{}, nil
	}

	queryMembers := make(types.QueryMembers, 0, len(resp.Entries))
	for _, entry := range resp.Entries {
		uid := uidFromEntry(entry)
		if uid == "" {
			continue
		}
		queryMembers = append(queryMembers, types.QueryMember{
			UID:     uid,
			Manager: entry.GetAttributeValue("manager"),
		})
	}

	log.Info("fetched query members")
	return queryMembers, nil
}

func uidFromEntry(entry *ldap.Entry) string {
	if entry == nil {
		return ""
	}
	uid := entry.GetAttributeValue("uid")
	if uid != "" {
		return uid
	}
	if dn, parseErr := ldap.ParseDN(entry.DN); parseErr == nil {
		return parseUIDFromDN(dn)
	}
	return ""
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

func (l *LDAPConn) BuildLDAPQueryString(ctx context.Context, query *v1alpha1.LDAPQuery) (string, error) {
	log := logger.Logger(ctx).WithField("build_ldap_query", "spec")
	log.WithField("query", query).Info("building LDAP query string from spec")

	if query == nil {
		return "", errors.New("ldap query is nil")
	}
	return buildQueryString(query, l.baseUserDN, 1)
}

func buildQueryString(query *v1alpha1.LDAPQuery, baseUserDN string, depth int) (string, error) {
	if depth > v1alpha1.MaxLDAPQueryDepth {
		return "", fmt.Errorf("ldap query nesting exceeds maximum depth of %d", v1alpha1.MaxLDAPQueryDepth)
	}
	if len(query.Filters) == 0 {
		return "", errors.New("filters are empty")
	}

	parts := make([]string, 0, len(query.Filters))
	for i, filter := range query.Filters {
		part, err := buildFilterItem(filter, baseUserDN, depth)
		if err != nil {
			return "", fmt.Errorf("filters[%d]: %w", i, err)
		}
		parts = append(parts, part)
	}

	op := strings.ToLower(strings.TrimSpace(query.Operator))
	switch op {
	case "and":
		return "(&" + strings.Join(parts, "") + ")", nil
	case "or":
		return "(|" + strings.Join(parts, "") + ")", nil
	default:
		return "", fmt.Errorf("unsupported operator %q", query.Operator)
	}
}

func buildFilterItem(filter v1alpha1.LDAPFilter, baseUserDN string, depth int) (string, error) {
	hasSimple := filter.Key != "" || filter.Criteria != "" || filter.Value != ""
	hasNested := filter.LDAPQuery != nil

	if hasSimple && hasNested {
		return "", errors.New("filter item cannot have both key/criteria/value and ldap_query")
	}

	if hasNested {
		return buildQueryString(&v1alpha1.LDAPQuery{
			Operator: filter.LDAPQuery.Operator,
			Filters:  filter.LDAPQuery.Filters,
		}, baseUserDN, depth+1)
	}

	return buildSimpleFilter(filter, baseUserDN)
}

func buildSimpleFilter(filter v1alpha1.LDAPFilter, baseUserDN string) (string, error) {
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
