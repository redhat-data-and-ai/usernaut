package ldap

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	ldapv3 "github.com/go-ldap/ldap/v3"
	v1alpha1 "github.com/redhat-data-and-ai/usernaut/api/v1alpha1"
)

type LDAP struct {
	Server           string   `yaml:"server"`
	BaseDN           string   `yaml:"baseDN"`
	BaseUserDN       string   `yaml:"baseUserDN"`
	UserDN           string   `yaml:"userDN"`
	UserSearchFilter string   `yaml:"userSearchFilter"`
	Attributes       []string `yaml:"attributes"`
}

type LDAPConnClient interface {
	IsClosing() bool
	Search(*ldapv3.SearchRequest) (*ldapv3.SearchResult, error)
	UnauthenticatedBind(username string) error
}

type closeableLDAPConnClient interface {
	LDAPConnClient
	Close() error
}

var dialLDAP = func(server string) (closeableLDAPConnClient, error) {
	return ldapv3.DialURL(server, ldapv3.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
}

type LDAPConn struct {
	mu               sync.RWMutex
	conn             LDAPConnClient
	connGeneration   uint64
	userDN           string
	baseDN           string
	baseUserDN       string
	server           string
	userSearchFilter string
	attributes       []string
}

type LDAPClient interface {
	GetUserLDAPData(ctx context.Context, userID string) (map[string]interface{}, error)
	GetBulkUserLDAPData(ctx context.Context, userIDs []string) (map[string]map[string]interface{}, error)
	GetQueryMembers(ctx context.Context, query string) ([]string, error)
	BuildLDAPQueryFromSpec(ctx context.Context, query *v1alpha1.LDAPQuery) (string, error)
	GetUserLDAPDataByEmail(ctx context.Context, email string) (map[string]interface{}, error)
}

// InitLdap initializes a connection to the LDAP server using the provided configuration.
func InitLdap(ldapConfig LDAP) (LDAPClient, error) {
	ldapConn, err := dialLDAP(ldapConfig.Server)
	if err != nil {
		return nil, err
	}

	// Perform anonymous bind (equivalent to ldapsearch -x)
	err = ldapConn.UnauthenticatedBind("")
	if err != nil {
		_ = ldapConn.Close()
		return nil, fmt.Errorf("failed to bind LDAP connection: %w", err)
	}

	return &LDAPConn{
		conn:             ldapConn,
		server:           ldapConfig.Server,
		userDN:           ldapConfig.UserDN,
		baseDN:           ldapConfig.BaseDN,
		baseUserDN:       ldapConfig.BaseUserDN,
		userSearchFilter: ldapConfig.UserSearchFilter,
		attributes:       ldapConfig.Attributes,
	}, nil
}

// getConn returns the underlying LDAP connection.
func (l *LDAPConn) getConn() LDAPConnClient {
	l.mu.RLock()
	conn := l.conn
	connGeneration := l.connGeneration
	if conn != nil && !conn.IsClosing() {
		l.mu.RUnlock()
		return conn
	}
	l.mu.RUnlock()

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.connGeneration != connGeneration {
		return l.conn
	}

	if l.conn != nil {
		newConn, err := dialLDAP(l.server)
		if err != nil {
			// Log the error and return the existing connection (or nil if no valid connection exists)
			fmt.Printf("Failed to re-establish LDAP connection: %v\n", err)
			return nil
		}
		// Perform anonymous bind (equivalent to ldapsearch -x)
		err = newConn.UnauthenticatedBind("")
		if err != nil {
			fmt.Printf("Failed to bind re-established LDAP connection: %v\n", err)
			_ = newConn.Close()
			return nil
		}
		l.conn = newConn
		l.connGeneration++
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
