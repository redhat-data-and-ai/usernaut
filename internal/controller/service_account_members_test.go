package controller

import (
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
)

func TestIgnoreServiceAccounts(t *testing.T) {
	ldapData := map[string]*structs.LDAPUser{
		"personone":      {UID: "personone", ObjectClass: "rhatPerson"},
		"serviceaccount-one": {UID: "serviceaccount-one", ObjectClass: "top rhatserviceaccount"},
	}
	members := []string{"personone", "serviceaccount-one", "missing"}
	log := logrus.NewEntry(logrus.New())

	got := ignoreServiceAccounts(t.Context(), members, ldapData, log)
	if len(got) != 2 || got[0] != "personone" || got[1] != "missing" {
		t.Fatalf("ignored list = %v", got)
	}

	kept := ignoreServiceAccounts(config.WithAllowServiceAccounts(t.Context()), members, ldapData, log)
	if len(kept) != len(members) {
		t.Fatalf("allowed list = %v", kept)
	}
}
