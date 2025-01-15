// Copyright 2025 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/utils/v2/env"
	"github.com/majewsky/schwift/v2"
	"github.com/majewsky/schwift/v2/gopherschwift"
	"github.com/sapcc/go-bits/osext"
)

// SwiftExporter handles exporting data to Swift
type SwiftExporter struct {
	container *schwift.Container
	format    string
	filename  string
	segSize   uint64
}

// initializeSwiftContainer creates and initializes a Swift container
func initializeSwiftContainer(ctx context.Context, provider *gophercloud.ProviderClient, containerName string) (*schwift.Container, error) {
	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: osext.GetenvOrDefault("OS_REGION_NAME", env.Getenv("OS_REGION_NAME")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Swift client: %w", err)
	}

	swiftAccount, err := gopherschwift.Wrap(client, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create swift account: %w", err)
	}

	container, err := swiftAccount.Container(containerName).EnsureExists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}

	return container, nil
}

// NewSwiftExporter creates a new SwiftExporter instance
func NewSwiftExporter(ctx context.Context, provider *gophercloud.ProviderClient, containerName, format, filename string, segmentSize uint64) (*SwiftExporter, error) {
	container, err := initializeSwiftContainer(ctx, provider, containerName)
	if err != nil {
		return nil, err
	}

	if filename == "" {
		filename = "hermes-export-" + time.Now().Format("2006-01-02-150405")
	}

	// Default segment size to 100MB if not specified
	if segmentSize == 0 {
		segmentSize = 100 * 1024 * 1024 // 100MB in bytes
	}

	return &SwiftExporter{
		container: container,
		format:    format,
		filename:  filename,
		segSize:   segmentSize,
	}, nil
}

// Upload uploads data to Swift, automatically handling large files through segmentation
func (e *SwiftExporter) Upload(ctx context.Context, reader io.Reader) error {
	filename := fmt.Sprintf("%s.%s", e.filename, e.format)
	obj := e.container.Object(filename)

	// Setup headers
	headers := make(schwift.Headers)
	headers.Set("Content-Type", getContentType(e.format))

	// Create segmentation options
	segmentOpts := schwift.SegmentingOptions{
		SegmentContainer: e.container,
		SegmentPrefix:    e.filename + "-segments/",
		Strategy:         schwift.StaticLargeObject,
	}

	// Create large object
	largeObject, err := obj.AsNewLargeObject(ctx, segmentOpts, nil)
	if err != nil {
		return fmt.Errorf("failed to create large object: %w", err)
	}

	// Upload the data
	segSize := e.segSize
	if segSize > uint64(math.MaxInt64) {
		return errors.New("segment size exceeds maximum int64 value")
	}
	if err := largeObject.Append(ctx, reader, int64(segSize), headers.ToOpts()); err != nil {
		return fmt.Errorf("failed to upload segments: %w", err)
	}

	// Write the manifest to complete the upload
	if err := largeObject.WriteManifest(ctx, nil); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// getContentType returns content type based on format
func getContentType(format string) string {
	switch format {
	case "json":
		return "application/json"
	case "csv":
		return "text/csv"
	case "yaml":
		return "application/x-yaml"
	default:
		return "application/octet-stream"
	}
}
