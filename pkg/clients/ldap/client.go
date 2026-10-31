package ldap

import (
	"context"
	"fmt"
	"net"
	"strings"
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
	BindUsername     string   `yaml:"bindUsername"`
	BindPassword     string   `yaml:"bindPassword"`

	// ServiceAccountObjectFilter is optional unless a namespace allows service accounts.
	ServiceAccountObjectFilter string `yaml:"serviceAccountObjectFilter"`
}

// Validate checks LDAP syntax when a service-account filter is configured.
func (l LDAP) Validate() error {
	if strings.TrimSpace(l.ServiceAccountObjectFilter) == "" {
		return nil
	}
	if _, err := ldapv3.CompileFilter(normalizeSearchFilter(l.ServiceAccountObjectFilter)); err != nil {
		return fmt.Errorf("invalid ldap.serviceAccountObjectFilter: %w", err)
	}
	return nil
}

type LDAPConnClient interface {
	IsClosing() bool
	Search(*ldapv3.SearchRequest) (*ldapv3.SearchResult, error)
	Bind(username, password string) error
	UnauthenticatedBind(username string) error
}

type closeableLDAPConnClient interface {
	LDAPConnClient
	Close() error
	SetTimeout(time.Duration)
}

// ldapBindTimeout bounds the bind in createConn, which runs under the write
// lock when getConn reconnects, so an unresponsive server cannot stall every
// caller. It does not apply to searches on the established connection.
const ldapBindTimeout = 30 * time.Second

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
	bindUsername     string
	bindPassword     string
	attributes       []string

	serviceAccountObjectFilter string
}

type LDAPClientConfig struct {
	server       string
	bindUsername string
	bindPassword string
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
	if err := ldapConfig.Validate(); err != nil {
		return nil, err
	}
	ldapClientConfig := LDAPClientConfig{
		server:       ldapConfig.Server,
		bindUsername: ldapConfig.BindUsername,
		bindPassword: ldapConfig.BindPassword,
	}

	conn, err := ldapClientConfig.createConn()
	if err != nil {
		return nil, err
	}

	return &LDAPConn{
		conn:             conn,
		server:           ldapConfig.Server,
		userDN:           ldapConfig.UserDN,
		baseDN:           ldapConfig.BaseDN,
		baseUserDN:       ldapConfig.BaseUserDN,
		userSearchFilter: ldapConfig.UserSearchFilter,
		attributes:       ldapConfig.Attributes,
		bindUsername:     ldapConfig.BindUsername,
		bindPassword:     ldapConfig.BindPassword,

		serviceAccountObjectFilter: strings.TrimSpace(ldapConfig.ServiceAccountObjectFilter),
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

	if l.connGeneration != connGeneration && l.conn != nil && !l.conn.IsClosing() {
		return l.conn
	}

	if l.conn != nil {
		ldapClientConfig := LDAPClientConfig{
			server:       l.server,
			bindUsername: l.bindUsername,
			bindPassword: l.bindPassword,
		}

		newConn, err := ldapClientConfig.createConn()
		if err != nil {
			// Log the error and return the existing connection (or nil if no valid connection exists)
			fmt.Printf("Failed to create LDAP connection: %v\n", err)
			return nil
		}
		l.conn = newConn
		l.connGeneration++
	}

	return l.conn
}

func (l *LDAPClientConfig) createConn() (LDAPConnClient, error) {
	newConn, err := dialLDAP(l.server)
	if err != nil {
		return nil, fmt.Errorf("failed to establish LDAP connection: %w", err)

	}

	newConn.SetTimeout(ldapBindTimeout)

	// Perform bind if username and password are set
	if strings.TrimSpace(l.bindUsername) != "" && strings.TrimSpace(l.bindPassword) != "" {
		err = newConn.Bind(l.bindUsername, l.bindPassword)
		if err != nil {
			_ = newConn.Close()
			return nil, fmt.Errorf("failed to bind LDAP connection: %w", err)
		}
	} else {
		err = newConn.UnauthenticatedBind(l.bindUsername)
		if err != nil {
			_ = newConn.Close()
			return nil, fmt.Errorf("failed to bind LDAP connection: %w", err)
		}
	}
	newConn.SetTimeout(0)
	return newConn, nil
}

// GetUserDN returns the user DN for the LDAP connection.
func (l *LDAPConn) GetUserDN() string {
	return l.userDN
}

// GetBaseDN returns the base DN for the LDAP connection.
func (l *LDAPConn) GetBaseDN() string {
	return l.baseDN
}
