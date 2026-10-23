package structs

import "testing"

func TestLDAPUserIsServiceAccount(t *testing.T) {
	if (*LDAPUser)(nil).IsServiceAccount() {
		t.Fatal("nil user")
	}
	if (&LDAPUser{ObjectClass: "top rhatPerson"}).IsServiceAccount() {
		t.Fatal("person")
	}
	if !(&LDAPUser{ObjectClass: "top rhatserviceaccount"}).IsServiceAccount() {
		t.Fatal("service account")
	}
	if !(&LDAPUser{ObjectClass: "top RHATSERVICEACCOUNT"}).IsServiceAccount() {
		t.Fatal("service account class is case-insensitive")
	}
}
