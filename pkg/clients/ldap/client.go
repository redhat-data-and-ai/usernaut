package ldap

import (
	"context"
	"fmt"
	"net"
	"strings"
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
}

type LDAPConnClient interface {
	IsClosing() bool
	Search(*ldapv3.SearchRequest) (*ldapv3.SearchResult, error)
	Bind(username, password string) error
	UnauthenticatedBind(username string) error
}

type LDAPConn struct {
	conn             LDAPConnClient
	userDN           string
	baseDN           string
	baseUserDN       string
	server           string
	userSearchFilter string
	bindUsername     string
	bindPassword     string
	attributes       []string
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
	}, nil
}

// getConn returns the underlying LDAP connection.
func (l *LDAPConn) getConn() LDAPConnClient {
	if l.conn != nil && l.conn.IsClosing() {
		ldapClientConfig := LDAPClientConfig{
			server:       l.server,
			bindUsername: l.bindUsername,
			bindPassword: l.bindPassword,
		}

		conn, err := ldapClientConfig.createConn()
		if err != nil {
			// Log the error and return the existing connection (or nil if no valid connection exists)
			fmt.Printf("Failed to create LDAP connection: %v\n", err)
			return nil
		}
		l.conn = conn
		return conn
	}

	return l.conn
}

func (l *LDAPClientConfig) createConn() (LDAPConnClient, error) {
	newConn, err := ldapv3.DialURL(l.server, ldapv3.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}))
	if err != nil {
		return nil, fmt.Errorf("Failed to establish LDAP connection: %v\n", err)

	}

	// Perform bind if username and password are set
	if strings.TrimSpace(l.bindUsername) != "" && strings.TrimSpace(l.bindPassword) != "" {
		err = newConn.Bind(l.bindUsername, l.bindPassword)
		if err != nil {
			_ = newConn.Close()
			return nil, fmt.Errorf("failed to bind LDAP connection: %v\n", err)
		}
	} else {
		err = newConn.UnauthenticatedBind(l.bindUsername)
		if err != nil {
			_ = newConn.Close()
			return nil, fmt.Errorf("failed to bind LDAP connection: %v\n", err)
		}
	}
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
