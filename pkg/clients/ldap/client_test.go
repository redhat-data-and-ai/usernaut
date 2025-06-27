package ldap

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitLdap_Error(t *testing.T) {
	LDAPConfig := LDAP{
		Server:           "ldap://ldap.com:389",
		BaseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		UserDN:           "uid=%s,ou=users,dc=example,dc=com",
		UserSearchFilter: "(objectClass=uid)",
		Attributes:       []string{"mail"},
	}

	_, err := InitLdap(LDAPConfig)
	assert.Error(t, err, "Expected error due to missing LDAP server connection")
}

func TestInitLdap_Success(t *testing.T) {
	addr, stop := startMockLDAPServer(t)
	defer stop()

	LDAPConfig := LDAP{
		// using a valid LDAP server for testing, reference: https://github.com/go-ldap/ldap/blob/master/v3/ldap_test.go#L13
		Server:           fmt.Sprintf("ldap://%s", addr),
		BaseDN:           "ou=adhoc,ou=managedGroups,dc=example,dc=com",
		UserDN:           "uid=%s,ou=users,dc=example,dc=com",
		UserSearchFilter: "(objectClass=uid)",
		Attributes:       []string{"mail"},
	}

	client, err := InitLdap(LDAPConfig)
	assert.NoError(t, err, "Expected successful LDAP client initialization")
	assert.NotNil(t, client, "Expected non-nil LDAP client")
}

// startMockLDAPServer starts a simple mock LDAP server for testing purposes.
func startMockLDAPServer(t *testing.T) (addr string, stop func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock LDAP server: %v", err)
	}
	done := make(chan struct{})
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					t.Logf("mock LDAP server accept error: %v", err)
					continue
				}
			}
			go func(c net.Conn) {
				defer func() {
					_ = c.Close()
				}()
				// Minimal LDAP handshake: just close after a short delay
				time.Sleep(100 * time.Millisecond)
			}(conn)
		}
	}()
	return ln.Addr().String(), func() {
		close(done)
		_ = ln.Close()
	}
}
