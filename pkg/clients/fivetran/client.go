package fivetran

import (
	"github.com/fivetran/go-fivetran"
	"github.com/redhat-data-and-ai/usernaut/pkg/clients"
)

type FivetranClient struct {
	fivetranClient *fivetran.Client
}

func NewClient(apiKey, apiSecret string) clients.Client {
	return &FivetranClient{
		fivetranClient: fivetran.New(apiKey, apiSecret),
	}
}
