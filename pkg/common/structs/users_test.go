package structs

import "testing"

func TestLDAPUserIsServiceAccount(t *testing.T) {
	for _, tc := range []struct {
		classes, configured string
		want                bool
	}{
		{"top person", "serviceAccount", false},
		{"top serviceAccount", "serviceAccount", true},
		{"top SERVICEACCOUNT", "serviceAccount", true},
		{"top notserviceAccount", "serviceAccount", false},
		{"top serviceAccount", "", false},
	} {
		if got := (&LDAPUser{ObjectClass: tc.classes}).IsServiceAccount(tc.configured); got != tc.want {
			t.Errorf("classes=%q configured=%q: got %v", tc.classes, tc.configured, got)
		}
	}
	if (*LDAPUser)(nil).IsServiceAccount("serviceAccount") {
		t.Fatal("nil user")
	}
}
