package model

import (
	"reflect"

	"github.com/atlanhq/atlan-go/atlan/model/structs"
)

// Unmarshal on assets changed the unmarshalling for the whole sdk asset structure

// Add Mutated assets for Response in Creation, Updation and Deletion
type MutatedAssets struct {
	TypeName            string        `json:"typeName"`
	Attributes          structs.Asset `json:"attributes"`
	Guid                string        `json:"guid"`
	Status              string        `json:"status"`
	DisplayText         string        `json:"displayText"`
	ClassificationNames []string      `json:"classificationNames"`
	MeaningNames        []string      `json:"meaningNames"`
	Meanings            []string      `json:"meanings"`
	IsIncomplete        bool          `json:"isIncomplete"`
	Labels              []string      `json:"labels"`
	CreatedBy           string        `json:"createdBy"`
	UpdatedBy           string        `json:"updatedBy"`
	CreateTime          int64         `json:"createTime"`
	UpdateTime          int64         `json:"updateTime"`
}

// Unmarshalling for assets from JSON
// Used in RetrieveMinimal Function

type MutatedEntities struct {
	// assets that were assets_updated. The detailed properties of the returned asset will vary based on
	// the type of asset, but listed in the example are the common set of properties across assets.
	UPDATE []*MutatedAssets `json:"UPDATE"`

	// assets that were created. The detailed properties of the returned asset will vary based on the
	// type of asset, but listed in the example are the common set of properties across assets.
	CREATE []*MutatedAssets `json:"CREATE"`

	// assets that were deleted. The detailed properties of the returned asset will vary based on the
	// type of asset, but listed in the example are the common set of properties across assets.
	DELETE []*MutatedAssets `json:"DELETE"`

	// assets that were partially updated. The detailed properties of the returned asset will
	// vary based on the type of asset, but listed in the example are the common set of properties across assets.
	PARTIAL_UPDATE []*MutatedAssets `json:"PARTIAL_UPDATE"`
}

type AssetMutationResponse struct {
	// Map of assigned unique identifiers for the changed assets.
	GuidAssignments map[string]string `json:"guidAssignments,omitempty"`

	// assets that were changed.
	MutatedEntities *MutatedEntities `json:"mutatedEntities"`

	// assets that were partially updated.
	PartialUpdatedEntities []*MutatedAssets `json:"partialUpdatedEntities,omitempty"`
}

func (amr *AssetMutationResponse) AssetsUpdated(assetType reflect.Type) []*MutatedAssets {
	if amr.MutatedEntities != nil && amr.MutatedEntities.UPDATE != nil {
		var assets []*MutatedAssets
		for _, asset := range amr.MutatedEntities.UPDATE {
			if reflect.TypeOf(asset).Elem() == assetType {
				assets = append(assets, asset)
			}
		}
		return assets
	}
	return []*MutatedAssets{}
}

func (amr *AssetMutationResponse) AssetsCreated(assetType reflect.Type) []*MutatedAssets {
	if amr.MutatedEntities != nil && amr.MutatedEntities.CREATE != nil {
		var assets []*MutatedAssets
		for _, asset := range amr.MutatedEntities.CREATE {
			if reflect.TypeOf(asset).Elem() == assetType {
				assets = append(assets, asset)
			}
		}
		return assets
	}
	return []*MutatedAssets{}
}

func (amr *AssetMutationResponse) AssetsDeleted(assetType reflect.Type) []*MutatedAssets {
	if amr.MutatedEntities != nil && amr.MutatedEntities.DELETE != nil {
		var assets []*MutatedAssets
		for _, asset := range amr.MutatedEntities.DELETE {
			if reflect.TypeOf(asset).Elem() == assetType {
				assets = append(assets, asset)
			}
		}
		return assets
	}
	return []*MutatedAssets{}
}

func (amr *AssetMutationResponse) AssetsPartiallyUpdated(assetType reflect.Type) []*MutatedAssets {
	if amr.MutatedEntities != nil && amr.MutatedEntities.PARTIAL_UPDATE != nil {
		var assets []*MutatedAssets
		for _, asset := range amr.MutatedEntities.PARTIAL_UPDATE {
			if reflect.TypeOf(asset).Elem() == assetType {
				assets = append(assets, asset)
			}
		}
		return assets
	}
	return []*MutatedAssets{}
}
