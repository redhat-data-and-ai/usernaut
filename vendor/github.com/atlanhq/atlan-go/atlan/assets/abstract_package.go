package assets

import (
	"encoding/json"
	"time"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// AbstractPackage represents a base package
type AbstractPackage struct {
	Parameters      []structs.NameValuePair
	CredentialsBody map[string]interface{}
	PackageName     string
	PackagePrefix   string
}

// NewAbstractPackage initializes an abstract package
func NewAbstractPackage(packageName, packagePrefix string) *AbstractPackage {
	return &AbstractPackage{
		PackageName:     packageName,
		PackagePrefix:   packagePrefix,
		Parameters:      []structs.NameValuePair{},
		CredentialsBody: map[string]interface{}{},
	}
}

func (p *AbstractPackage) ToWorkflow() *structs.Workflow {
	metadata := p.GetMetadata()

	spec := structs.WorkflowSpec{
		Entrypoint: structs.StringPtr("main"),
		Templates: []structs.WorkflowTemplate{
			{
				Name: "main",
				DAG: structs.WorkflowDAG{
					Tasks: []structs.WorkflowTask{
						{
							Name: "run",
							Arguments: structs.WorkflowParameters{
								Parameters: p.Parameters,
							},
							TemplateRef: structs.WorkflowTemplateRef{
								Name:         p.PackagePrefix,
								Template:     "main",
								ClusterScope: true,
							},
						},
					},
				},
			},
		},
		WorkflowMetadata: metadata,
	}

	var payload []structs.PackageParameter
	if len(p.CredentialsBody) > 0 {
		credJSON, _ := json.Marshal(p.CredentialsBody)
		payload = append(payload, structs.PackageParameter{
			Parameter: "credentialGuid",
			Type:      "credential",
			Body:      credJSON,
		})
	}

	return &structs.Workflow{
		Metadata: metadata,
		Spec:     &spec,
		Payload:  payload,
	}
}

// GetMetadata should be implemented by subclasses
func (p *AbstractPackage) GetMetadata() *structs.WorkflowMetadata {
	// Default (empty) metadata implementation, to be overridden by child structs
	return &structs.WorkflowMetadata{}
}

// AbstractMiner represents a base miner package
type AbstractMiner struct {
	*AbstractPackage
	Epoch int64
}

// NewAbstractMiner initializes an abstract miner
func NewAbstractMiner(connectionQualifiedName, packageName, packagePrefix string) *AbstractMiner {
	epoch := time.Now().Unix()
	packageInstance := NewAbstractPackage(packageName, packagePrefix)
	packageInstance.Parameters = append(packageInstance.Parameters, structs.NameValuePair{
		Name:  "connection-qualified-name",
		Value: connectionQualifiedName,
	})

	return &AbstractMiner{
		AbstractPackage: packageInstance,
		Epoch:           epoch,
	}
}
