// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package index

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"cloudeng.io/errors"
	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/gleansdk"
)

// BulkFlags represents the flags to the bulk indexing command.
type BulkFlags struct {
	config.FileFlags
	UploadID      string `subcmd:"upload-id,upload,id to use for this bulk upload"`
	ForceRestart  bool   `subcmd:"force-restart,false,restart the bulk upload"`
	ForceDeletion bool   `subcmd:"force-sync-deletion,false,synchronously delete stale documents on upload of last bulk indexing batch"`
}

// Bulk indexes a datasource in bulk mode.
func Bulk(ctx context.Context, gleanConfig config.GleanConfig, fv *BulkFlags, datasource string) error {
	cfg, err := config.DatasourceForName(fv.ConfigFile, datasource)
	if err != nil {
		return err
	}
	cachePath := os.ExpandEnv(cfg.Cache.Path)
	if len(cachePath) == 0 {
		return fmt.Errorf("no path specified for the cache to be indexed")
	}

	ctx, client, err := gleanConfig.NewAPIClient(ctx, cfg.GleanInstance)
	if err != nil {
		return err
	}

	cnvs, err := NewConverters(cfg)
	if err != nil {
		return err
	}

	size := cfg.Index.ReaddirEntries
	if size == 0 {
		size = 100
	}
	reqCh := make(chan Request, size)

	forceRestart, forceDeletion := cfg.ForceRestart, cfg.ForceDeletion
	if fv.ForceDeletion {
		forceDeletion = true
	}

	if fv.ForceRestart {
		forceRestart = true
	}

	walker := NewWalker(cnvs, cfg.Index.ReaddirEntries, reqCh)
	indexer := NewBulkIndexer(client, cfg,
		WithForceDelete(forceDeletion),
		WithForceRestart(forceRestart),
		WithBulkID(fv.UploadID))

	errs := &errors.M{}
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		errs.Append(walker.Run(ctx, cachePath))
		close(reqCh)
		wg.Done()
	}()

	go func() {
		errs.Append(indexer.Run(ctx, reqCh))
		wg.Done()

	}()
	wg.Wait()
	return errs.Err()
}

// BulkIndexOption represents an option to NewBulkIndexer.
type BulkIndexOption func(o *bulkOptions)

type bulkOptions struct {
	forceDeletion bool
	forceRestart  bool
	id            string
}

// WithForceDelete disables the deletion of too many documents test.
func WithForceDelete(forceDelete bool) BulkIndexOption {
	return func(o *bulkOptions) {
		o.forceDeletion = forceDelete
	}
}

// WithForceRestart sets the force restart options.
func WithForceRestart(forceRestart bool) BulkIndexOption {
	return func(o *bulkOptions) {
		o.forceRestart = forceRestart
	}
}

// WithBulkID specifies a custom id to use for the bulk upload. If one
// is not specified the current date and time are used.
func WithBulkID(id string) BulkIndexOption {
	return func(o *bulkOptions) {
		o.id = id
	}
}

// BulkIndexer represents a bulk indexer.
type BulkIndexer struct {
	bulkOptions
	id     string
	cfg    config.Datasource
	client *gleansdk.APIClient
}

// NewBulkIndexer creates a new bulk indexer.
func NewBulkIndexer(client *gleansdk.APIClient, datasource config.Datasource, opts ...BulkIndexOption) *BulkIndexer {
	b := &BulkIndexer{
		id:     time.Now().Round(0).String(),
		cfg:    datasource,
		client: client}
	for _, fn := range opts {
		fn(&b.bulkOptions)
	}
	return b
}

// Run runs the a bulk index operation, receiving requests to be indexed over
// the specified channel.
func (b *BulkIndexer) Run(ctx context.Context, ch <-chan Request) error {
	var (
		firstPage = true
		indexed   = 0
		duration  time.Duration
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req, ok := <-ch:
			if !ok {
				if err := b.sendLastPage(ctx); err != nil {
					return err
				}
				if duration > 0 && indexed > 0 {
					avg := time.Duration(int64(duration) / int64(indexed))
					fmt.Printf("indexed: all # docs: % 5v docs in % 8v, (avg: %8v)\n", indexed, duration, avg)
				}
				return nil
			}
			if len(req.Documents) == 0 {
				continue
			}
			bulkReq := b.createBulkIndexReq(req.Documents)
			bulkReq.SetIsFirstPage(firstPage)
			if firstPage {
				bulkReq.SetForceRestartUpload(b.forceRestart)
			}
			bulkReq.SetIsLastPage(req.LastPage)
			bulkReq.SetUploadId(b.id)
			bulkReq.SetDisableStaleDocumentDeletionCheck(b.forceDeletion)
			reqStart := time.Now()
			resp, err := b.client.DocumentsApi.BulkindexdocumentsPost(ctx).BulkIndexDocumentsRequest(bulkReq).Execute()
			if err := handleHTTPError(resp, err); err != nil {
				return err
			}
			took := time.Since(reqStart)
			duration += took
			indexed += len(req.Documents)
			avg := time.Duration(int64(duration) / int64(indexed))
			fmt.Printf("indexed: total # docs: % 5v, per req # docs: % 3v in % 8v (avg: %8v)\n", indexed, len(bulkReq.Documents), took, avg)
			firstPage = false
		}
	}
}

func (b *BulkIndexer) createBulkIndexReq(gdocs []*gleansdk.DocumentDefinition) gleansdk.BulkIndexDocumentsRequest {
	var req gleansdk.BulkIndexDocumentsRequest
	req.Datasource = b.cfg.CustomDatasourceConfig.Name
	for _, gd := range gdocs {
		req.Documents = append(req.Documents, *gd)
	}
	return req
}

func (b *BulkIndexer) sendLastPage(ctx context.Context) error {
	bulkReq := gleansdk.BulkIndexDocumentsRequest{}
	bulkReq.SetDatasource(b.cfg.CustomDatasourceConfig.GetName())
	bulkReq.SetIsFirstPage(false)
	bulkReq.SetIsLastPage(true)
	bulkReq.SetUploadId(b.id)
	bulkReq.Documents = []gleansdk.DocumentDefinition{}
	resp, err := b.client.DocumentsApi.BulkindexdocumentsPost(ctx).BulkIndexDocumentsRequest(bulkReq).Execute()
	if err := handleHTTPError(resp, err); err != nil {
		return fmt.Errorf("last page request: %v", err)
	}
	return nil
}
