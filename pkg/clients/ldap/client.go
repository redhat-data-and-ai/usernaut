package ldap

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/go-ldap/ldap/v3"
)

type LDAP struct {
	Server           string   `yaml:"server"`
	BaseDN           string   `yaml:"baseDN"`
	UserDN           string   `yaml:"userDN"`
	UserSearchFilter string   `yaml:"userSearchFilter"`
	Attributes       []string `yaml:"attributes"`
}

type LDAPConnClient interface {
	IsClosing() bool
	Search(*ldap.SearchRequest) (*ldap.SearchResult, error)
}

type LDAPConn struct {
	conn             LDAPConnClient
	userDN           string
	baseDN           string
	server           string
	userSearchFilter string
	attributes       []string
}

type LDAPClient interface {
	GetUserLDAPData(ctx context.Context, userID string) (map[string]interface{}, error)
}

// InitLdap initializes a connection to the LDAP server using the provided configuration.
func InitLdap(ldapConfig LDAP) (LDAPClient, error) {
	ldapConn, err := ldap.DialURL(ldapConfig.Server, ldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, err
	}

	return &LDAPConn{
		conn:             ldapConn,
		server:           ldapConfig.Server,
		userDN:           ldapConfig.UserDN,
		baseDN:           ldapConfig.BaseDN,
		userSearchFilter: ldapConfig.UserSearchFilter,
		attributes:       ldapConfig.Attributes,
	}, nil
}

// getConn returns the underlying LDAP connection.
func (l *LDAPConn) getConn() LDAPConnClient {
	if l.conn != nil && l.conn.IsClosing() {
		newConn, err := ldap.DialURL(l.server, ldap.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
		if err != nil {
			// Log the error and return the existing connection (or nil if no valid connection exists)
			fmt.Printf("Failed to re-establish LDAP connection: %v\n", err)
			return nil
		}
		l.conn = newConn
	}

	return l.conn
}

// GetUserDN returns the user DN for the LDAP connection.
func (l *LDAPConn) GetUserDN() string {
	return l.userDN
}

// GetBaseDN returns the base DN for the LDAP connection.
func (l *LDAPConn) GetBaseDN() string {
	return l.baseDN
}
