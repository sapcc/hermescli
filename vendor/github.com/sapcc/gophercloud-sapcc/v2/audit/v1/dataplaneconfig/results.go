// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package dataplaneconfig

import (
	"time"

	"github.com/gophercloud/gophercloud/v2"
)

// DataplaneConfig holds the per-project dataplane routing configuration
// returned by the Hermes API.
type DataplaneConfig struct {
	ProjectID    string    `json:"project_id"`
	Enabled      bool      `json:"enabled"`
	TargetBucket string    `json:"target_bucket,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
	UpdatedBy    string    `json:"updated_by"`
}

// GetResult represents the result of a Get operation.
type GetResult struct {
	gophercloud.Result
}

// Extract interprets a GetResult as a DataplaneConfig.
func (r GetResult) Extract() (*DataplaneConfig, error) {
	var s DataplaneConfig
	err := r.ExtractInto(&s)
	return &s, err
}

// ExtractInto is used by Extract.
func (r GetResult) ExtractInto(v any) error {
	return r.ExtractIntoStructPtr(v, "")
}

// PutResult represents the result of a Put operation.
type PutResult struct {
	gophercloud.Result
}

// Extract interprets a PutResult as a DataplaneConfig.
func (r PutResult) Extract() (*DataplaneConfig, error) {
	var s DataplaneConfig
	err := r.ExtractInto(&s)
	return &s, err
}

// ExtractInto is used by Extract.
func (r PutResult) ExtractInto(v any) error {
	return r.ExtractIntoStructPtr(v, "")
}

// DeleteResult represents the result of a Delete operation.
type DeleteResult struct {
	gophercloud.ErrResult
}
