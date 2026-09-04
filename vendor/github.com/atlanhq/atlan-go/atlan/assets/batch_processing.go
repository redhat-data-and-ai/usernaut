package assets

import (
	"errors"

	"github.com/atlanhq/atlan-go/atlan"
	"github.com/atlanhq/atlan-go/atlan/model"
)

// FailedBatch is an internal struct to capture batch failures.
type FailedBatch struct {
	FailedAssets  []AtlanObject // The list of assets that failed during processing.
	FailureReason error         // The reason for the batch failure.
}

// NewFailedBatch creates a new FailedBatch instance.
func NewFailedBatch(failedAssets []AtlanObject, failureReason error) (*FailedBatch, error) {
	if failureReason == nil {
		return nil, errors.New("failure reason cannot be nil")
	}

	return &FailedBatch{
		FailedAssets:  failedAssets,
		FailureReason: failureReason,
	}, nil
}

// GetFailedAssets returns the list of failed assets.
func (fb *FailedBatch) GetFailedAssets() []AtlanObject {
	return fb.FailedAssets
}

// GetFailureReason returns the reason for the batch failure.
func (fb *FailedBatch) GetFailureReason() error {
	return fb.FailureReason
}

// Batch is a utility class for managing bulk updates in batches.
type Batch struct {
	client                 *AtlanClient
	maxSize                int
	replaceAtlanTags       bool
	customMetadataHandling atlan.CustomMetadataHandling
	captureFailures        bool
	batch                  []AtlanObject
	failures               []FailedBatch
	created                []*model.MutatedAssets
	updated                []*model.MutatedAssets
}

// NewBatch creates a new Batch for managing bulk updates.
func NewBatch(client *AtlanClient, maxSize int, replaceAtlanTags bool, customMetadataHandling atlan.CustomMetadataHandling, captureFailures bool) *Batch {
	return &Batch{
		client:                 client,
		maxSize:                maxSize,
		replaceAtlanTags:       replaceAtlanTags,
		customMetadataHandling: customMetadataHandling,
		captureFailures:        captureFailures,
		batch:                  []AtlanObject{},
		failures:               []FailedBatch{},
		created:                []*model.MutatedAssets{},
		updated:                []*model.MutatedAssets{},
	}
}

// Failures returns a list of FailedBatch objects containing information about any failed batches.
func (b *Batch) Failures() []FailedBatch {
	return b.failures
}

// Created returns a list of Assets that were created.
func (b *Batch) Created() []*model.MutatedAssets {
	return b.created
}

// Updated returns a list of Assets that were updated.
func (b *Batch) Updated() []*model.MutatedAssets {
	return b.updated
}

/* In Case, We want a list of assets dereferenced with attributes

// Created returns a list of Assets that were created (dereferenced).
func (b *Batch) Created() []model.MutatedAssets {
	// Dereference the pointers to get the actual values.
	dereferencedCreated := make([]model.MutatedAssets, len(b.created))
	for i, asset := range b.created {
		dereferencedCreated[i] = *asset // Dereference the pointer to get the actual value
	}
	return dereferencedCreated
}

// Updated returns a list of Assets that were updated (dereferenced).
func (b *Batch) Updated() []model.MutatedAssets {
	// Dereference the pointers to get the actual values.
	dereferencedUpdated := make([]model.MutatedAssets, len(b.updated))
	for i, asset := range b.updated {
		dereferencedUpdated[i] = *asset // Dereference the pointer to get the actual value
	}
	return dereferencedUpdated
}

*/

// Add adds an asset to the batch and processes it if the batch size is reached.
func (b *Batch) Add(asset AtlanObject) error {
	b.batch = append(b.batch, asset)
	if len(b.batch) >= b.maxSize {
		_, err := b.Flush()
		return err
	}
	return nil
}

// process checks if the batch size is reached and flushes the batch if needed.
func (b *Batch) process() (*model.AssetMutationResponse, error) {
	if len(b.batch) == b.maxSize {
		return b.Flush()
	}
	return nil, nil
}

// Flush sends the current batch to the Save function and clears the batch.
func (b *Batch) Flush() (*model.AssetMutationResponse, error) {
	if len(b.batch) == 0 {
		return nil, nil // No assets to process
	}

	response, err := Save(b.batch...)
	if err != nil {
		if b.captureFailures {
			b.failures = append(b.failures, FailedBatch{
				FailedAssets:  b.batch,
				FailureReason: err,
			})
		} else {
			return nil, err
		}
	}

	if response != nil {
		b.trackResponse(response)
	}

	b.batch = []AtlanObject{} // Clear the batch after processing
	return response, nil
}

// trackResponse processes the response and updates the created and updated assets.
func (b *Batch) trackResponse(response *model.AssetMutationResponse) {
	if response.MutatedEntities != nil {
		for _, asset := range response.MutatedEntities.CREATE {
			b.track(&b.created, asset)
		}
		for _, asset := range response.MutatedEntities.UPDATE {
			b.track(&b.updated, asset)
		}
	}
}

// track adds an asset to the tracker.
func (b *Batch) track(tracker *[]*model.MutatedAssets, asset *model.MutatedAssets) {
	*tracker = append(*tracker, asset)
}
