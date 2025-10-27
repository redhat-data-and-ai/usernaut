/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gitlab

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/config"
	"github.com/redhat-data-and-ai/usernaut/pkg/request"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// NewClient creates a GitlabClient with a heimdall-backed HTTP client and passes it to the SDK.
func NewClient(
	gitlabAppConfig map[string]interface{},
	dependsOn config.Dependant,
	poolCfg httpclient.ConnectionPoolConfig,
	hystrixCfg httpclient.HystrixResiliencyConfig,
) (*GitlabClient, error) {
	gitlabConfig := GitlabConfig{}
	if err := utils.MapToStruct(gitlabAppConfig, &gitlabConfig); err != nil {
		return nil, err
	}

	if gitlabConfig.URL == "" || gitlabConfig.Token == "" {
		return nil, fmt.Errorf("missing required connection parameters for gitlab backend")
	}

	baseUrl := fmt.Sprintf("%s/api/v4", gitlabConfig.URL)
	gitlabConfig.URL = baseUrl

	// Gitlab SDK Client
	client, err := gitlab.NewClient(gitlabConfig.Token, gitlab.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}

	// Heimdall Client to initiate LDAP sync request
	heimdallClient, err := httpclient.InitializeClient(
		"gitlab",
		poolCfg,
		hystrixCfg,
		heimdall.NewRetrier(heimdall.NewConstantBackoff(100*time.Millisecond, 50*time.Millisecond)), 3,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	dependantExists := false
	if dependsOn.Name != "" || dependsOn.Type != "" {
		dependantExists = true
	}

	return &GitlabClient{
		gitlabClient:    client,
		gitlabConfig:    &gitlabConfig,
		dependantExists: dependantExists,
		httpClient:      heimdallClient,
	}, nil
}

func (g *GitlabClient) SetLdapSync(ldapSync bool, cn string) {
	g.ldapSync = ldapSync
	g.cn = cn
}

func (g *GitlabClient) sendLdapSyncRequest(ctx context.Context) ([]byte, int, error) {
	url := fmt.Sprintf("%s/groups/%d/ldap_sync", g.gitlabConfig.URL, g.gitlabConfig.ParentGroupId)
	requestBody := []byte{}
	request, err := request.NewRequest(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return nil, 0, err
	}
	request.SetHeaders(map[string]string{
		"Authorization": "Bearer " + g.gitlabConfig.Token,
	})
	return request.MakeRequest(g.httpClient, "backend.gitlab.InitiateLdapSync", "gitlab")
}
