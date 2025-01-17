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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/cheggaaa/pb/v3"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/utils/v2/env"
	"github.com/majewsky/schwift/v2"
	"github.com/majewsky/schwift/v2/gopherschwift"
)

// ExportFile represents a file to be exported to Swift storage
type ExportFile struct {
	Format      string
	FileName    string
	SegmentSize uint64
	Contents    io.Reader
}

func (f ExportFile) UploadTo(ctx context.Context, container *schwift.Container) error {
	if reader, ok := f.Contents.(*bytes.Buffer); ok {
		dataSize := float64(reader.Len()) / 1024 / 1024 // Convert to MB
		fmt.Fprintf(os.Stderr, "Uploading %.1fMB to Swift...\n", dataSize)

		// Create upload progress bar
		uploadBar := pb.Full.Start64(int64(reader.Len()))
		uploadBar.Set(pb.Bytes, true)
		uploadBar.SetWidth(80)
		defer uploadBar.Finish()

		// Wrap the buffer in a progress reader
		f.Contents = &ProgressReader{
			Reader: reader,
			Bar:    uploadBar,
		}
	}

	filename := fmt.Sprintf("%s.%s", f.FileName, f.Format)
	obj := container.Object(filename)

	// Setup headers
	headers := make(schwift.Headers)
	headers.Set("Content-Type", getContentType(f.Format))

	// Create segmentation options
	segmentOpts := schwift.SegmentingOptions{
		SegmentContainer: container,
		SegmentPrefix:    f.FileName + "-segments/",
		Strategy:         schwift.StaticLargeObject,
	}

	// Create large object
	largeObject, err := obj.AsNewLargeObject(ctx, segmentOpts, nil)
	if err != nil {
		return fmt.Errorf("failed to create large object: %w", err)
	}

	// Upload the data
	segSize := f.SegmentSize
	if segSize > uint64(math.MaxInt64) {
		return errors.New("segment size exceeds maximum int64 value")
	}
	if err := largeObject.Append(ctx, f.Contents, int64(segSize), headers.ToOpts()); err != nil {
		return fmt.Errorf("failed to upload segments: %w", err)
	}

	// Write the manifest to complete the upload
	if err := largeObject.WriteManifest(ctx, nil); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// InitializeSwiftContainer creates and initializes a Swift container
func InitializeSwiftContainer(ctx context.Context, provider *gophercloud.ProviderClient, containerName string) (*schwift.Container, error) {
	client, err := openstack.NewObjectStorageV1(provider, gophercloud.EndpointOpts{
		Region: env.Getenv("OS_REGION_NAME"),
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

// ProgressReader wraps an io.Reader to update a progress bar
type ProgressReader struct {
	Reader io.Reader
	Bar    *pb.ProgressBar
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	if n > 0 {
		pr.Bar.Add(n)
	}
	return
}
