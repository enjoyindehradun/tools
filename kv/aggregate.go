package kv

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cdnjs/tools/compress"
	"github.com/cdnjs/tools/util"

	"github.com/cdnjs/tools/packages"
)

// UpdateAggregatedMetadata updates a package's KV entry for aggregated metadata.
func UpdateAggregatedMetadata(ctx context.Context, pckg *packages.Package, newAssets []packages.Asset) ([]string, error) {
	aggPckg, err := getAggregatedMetadata(*pckg.Name)
	if err != nil {
		switch err.(type) {
		case KeyNotFoundError:
			{
				// key not found (new package)
				util.Debugf(ctx, "KV key `%s` not found, inserting aggregated metadata...\n", *pckg.Name)
				pckg.Assets = newAssets
			}
		default:
			{
				return nil, err
			}
		}
	} else {
		util.Debugf(ctx, "Aggregated metadata for `%s` found. Updating aggregated metadata...\n", *pckg.Name)
		pckg.Assets = append(aggPckg.Assets, newAssets...)
	}

	return writeAggregatedMetadata(ctx, pckg)
}

// Reads an aggregated metadata entry in KV, ungzipping it and
// unmarshalling it into a *packages.Package.
func getAggregatedMetadata(key string) (*packages.Package, error) {
	gzipBytes, err := Read(key, aggregatedMetadataNamespaceID)

	if err != nil {
		return nil, err
	}

	// unmarshal and ungzip
	var p packages.Package
	util.Check(json.Unmarshal(compress.UnGzip(gzipBytes), &p))

	return &p, nil
}

// Writes an aggregated metadata entry to KV, gzipping the bytes.
func writeAggregatedMetadata(ctx context.Context, p *packages.Package) ([]string, error) {
	// marshal package into JSON
	v, err := p.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal KV package JSON: %s", *p.Name)
	}

	// gzip the bytes
	req := &writeRequest{
		key:   *p.Name,
		value: compress.Gzip9Native(v),
	}

	// write aggregated to KV
	return encodeAndWriteKVBulk(ctx, []*writeRequest{req}, aggregatedMetadataNamespaceID)
}
