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
	"fmt"
	"time"

	"github.com/gojek/heimdall/v7"
	"github.com/redhat-data-and-ai/usernaut/pkg/request/httpclient"
	"github.com/redhat-data-and-ai/usernaut/pkg/utils"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// NewClient creates a GitlabClient with a heimdall-backed HTTP client and passes it to the SDK.
func NewClient(
	gitlabAppConfig map[string]interface{},
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
	client, err := gitlab.NewClient(gitlabConfig.Token, gitlab.WithBaseURL(baseUrl))
	if err != nil {
		return nil, err
	}

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

	gitlabConfig.URL = baseUrl
	return &GitlabClient{
		gitlabClient: client,
		ldapSync:     false,
		gitlabConfig: &gitlabConfig,
		httpClient:   heimdallClient,
	}, nil
}

func (g *GitlabClient) SetLdapSync(ldapSync bool, cn string) {
	g.ldapSync = ldapSync
	g.cn = cn
}
