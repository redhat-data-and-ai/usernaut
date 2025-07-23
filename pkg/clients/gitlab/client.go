package gitlab

import (
	"fmt"

	"github.com/redhat-data-and-ai/usernaut/pkg/common/structs"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabClient struct {
	gitlabClient  *gitlab.Client
	ldapSync      bool
	externalTeams map[string]*structs.Team // Teams from external backends (like Rover)
}

type GitlabConfig struct {
	URL      string `json:"url"`
	Token    string `json:"token"`
	LdapSync bool   `json:"ldap_sync"`
}

func NewClient(gitlabAppConfig map[string]interface{}) (*GitlabClient, error) {
	gitlabConfig := GitlabConfig{}
	if err := utils.MapToStruct(gitlabAppConfig, &gitlabConfig); err != nil {
		return nil, err
	}

	if gitlabConfig.URL == "" || gitlabConfig.Token == "" {
		return nil, fmt.Errorf("missing required connection parameters for gitlab backend")
	}

	baseUrl := fmt.Sprintf("%s/api/v4", gitlabConfig.URL)
	client, err := gitlab.NewClient(gitlabConfig.Token, gitlab.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}

	return &GitlabClient{
		gitlabClient: client,
		ldapSync:     gitlabConfig.LdapSync,
	}, nil
}

// SetExternalTeams allows setting teams from external backends (like Rover)
func (g *GitlabClient) SetExternalTeams(teams map[string]*structs.Team) {
	if g.externalTeams == nil {
		g.externalTeams = make(map[string]*structs.Team)
	}
	g.externalTeams = teams
}
