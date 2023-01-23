// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package crawl

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloudeng.io/aws/awsconfig"
	"cloudeng.io/aws/s3fs"
	"cloudeng.io/errors"
	"cloudeng.io/file/crawl"
	"cloudeng.io/file/crawl/outlinks"
	"cloudeng.io/file/download"
	"cloudeng.io/path"
	"cloudeng.io/path/cloudpath"
	"cloudeng.io/sync/errgroup"
	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/glean/gleancli/encoding"
)

type Flags struct {
	config.FileFlags
	Outlinks bool `subcmd:"outlinks,false,display extracted outlinks"`
	Progress bool `subcmd:"progress,true,'display progress of downloads'"`
}

func Run(ctx context.Context, gleanConfig config.GleanConfig, fv *Flags, datasource string) error {
	cfg, err := config.DatasourceForName(fv.ConfigFile, datasource)
	if err != nil {
		return err
	}
	if len(cfg.Cache.Path) == 0 {
		return fmt.Errorf("no path specified for the cache to stored downloaded files")
	}

	cachePath := os.ExpandEnv(cfg.Cache.Path)
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failedto ensure that %v exists: %v", cfg.Cache.Path, err)
	}
	for _, crawl := range cfg.Crawls {
		if len(crawl.CachePrefix) == 0 {
			continue
		}
		crawlCache := filepath.Join(cachePath, crawl.CachePrefix)
		if crawl.CacheClearBeforeCrawl {
			if err := os.RemoveAll(crawlCache); err != nil {
				return fmt.Errorf("failed to remove %v: %v", cachePath, err)
			}
		}
	}

	g := errgroup.T{}
	for _, crawl := range cfg.Crawls {
		crawlCache := filepath.Join(cachePath, crawl.CachePrefix)
		crawler := crawler{
			name:       crawl.Name,
			cachePath:  crawlCache,
			Flags:      fv,
			Crawl:      crawl,
			Datasource: cfg,
		}
		g.Go(func() error {
			return crawler.run(ctx)
		})
	}
	return g.Wait()
}

type crawler struct {
	*Flags
	name      string
	cachePath string
	config.Crawl
	config.Datasource
}

func (c *crawler) run(ctx context.Context) error {

	requests, err := c.createRequests(ctx)
	if err != nil {
		return err
	}

	dlFactory := DownloaderFactory{Download: c.Download}
	if c.Progress {
		go dlFactory.HandleProgress(ctx, c.Name, make(chan download.Progress, 100))
	}

	reqChanSize := depthOrDefault(0, c.Download.RequestChanSizes, c.Download.DefaultRequestChanSize)
	craweldChanSize := depthOrDefault(0, c.Download.CrawledChanSizes, c.Download.DefaultCrawledChanSize)

	reqCh := make(chan download.Request, reqChanSize)
	crawledCh := make(chan crawl.Crawled, craweldChanSize)

	extractors, err := c.extractorsFromConfig()
	if err != nil {
		return err
	}
	extractorErrCh := make(chan outlinks.Errors, 100)

	crawler := crawl.New(crawl.WithNumExtractors(c.Extractors.Concurrency),
		crawl.WithCrawlDepth(c.Depth))

	linkProcessor := &outlinks.RegexpProcessor{
		NoFollow: c.Crawl.NoFollowRules,
		Follow:   c.Crawl.FollowRules,
		Rewrite:  c.Crawl.RewriteRules,
	}
	if err := linkProcessor.Compile(); err != nil {
		return fmt.Errorf("failed to compile link processing rules: %v", err)
	}
	extractor := outlinks.NewExtractors(extractorErrCh, linkProcessor, extractors...)

	var errs errors.M
	var wg sync.WaitGroup
	wg.Add(3)

	go func(ch chan crawl.Crawled) {
		errs.Append(c.saveCrawled(ctx, c.name, ch))
		wg.Done()
	}(crawledCh)

	go func() {
		errs.Append(crawler.Run(ctx, dlFactory.New, extractor, reqCh, crawledCh))
		wg.Done()
	}()

	go func() {
		defer wg.Done()
		defer close(reqCh)
		for _, req := range requests {
			select {
			case <-ctx.Done():
				errs.Append(ctx.Err())
				return
			case reqCh <- req:
			}
		}
	}()

	go func() {
		for err := range extractorErrCh {
			if len(err.Errors) > 0 {
				fmt.Printf("extractor error: %v\n", err)
			}
		}
	}()

	wg.Wait()
	close(extractorErrCh)
	return errs.Err()
}

func (c crawler) saveCrawled(ctx context.Context, prefix string, crawledCh chan crawl.Crawled) error {
	sharder := path.NewSharder(path.WithSHA1PrefixLength(3))

	for crawled := range crawledCh {
		if c.Outlinks {
			for _, req := range crawled.Outlinks {
				fmt.Printf("%v\n", strings.Join(crawled.Request.Names(), " "))
				for _, name := range req.Names() {
					fmt.Printf("\t-> %v\n", name)
				}
			}
		}
		for _, dld := range crawled.Downloads {
			if dld.Err != nil {
				fmt.Printf("download error: %v: %v\n", dld.Name, dld.Err)
				continue
			}
			prefix, suffix := sharder.Assign(prefix + dld.Name)
			path := filepath.Join(c.cachePath, prefix, suffix)
			err := encoding.WriteDownload(filepath.Join(c.cachePath, prefix), suffix, dld)
			if err != nil {
				fmt.Printf("failed to write: %v as %v: %v\n", dld.Name, path, err)
				continue
			}
			fmt.Printf("%v -> %v\n", dld.Name, path)
		}
	}
	return nil
}

func (c crawler) extractorsFromConfig() ([]outlinks.Extractor, error) {
	exts := []outlinks.Extractor{}
	for _, ext := range c.Extractors.Use {
		switch ext {
		case "text/html":
			exts = append(exts, outlinks.NewHTML())
		default:
			return nil, fmt.Errorf("unsupported extractor %q", ext)
		}
	}
	return exts, nil
}

func (c crawler) createRequests(ctx context.Context) ([]download.Request, error) {
	requests := []download.Request{}
	reqs, err := c.requestsForSeeds(ctx)
	if err != nil {
		return nil, err
	}
	requests = append(requests, reqs...)
	return requests, nil
}

func (c crawler) requestsForSeeds(ctx context.Context) ([]download.Request, error) {
	matches := map[string][]cloudpath.Match{}
	for _, seed := range c.Seeds {
		match := cloudpath.DefaultMatchers.Match(seed)
		if len(match.Matched) == 0 {
			return nil, fmt.Errorf("failed to identify scheme for seed %q", seed)
		}
		scheme := match.Scheme
		matches[scheme] = append(matches[scheme], match)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no valid filenames/URIs found in %v", c.Seeds)
	}
	requests := []download.Request{}
	for scheme, matched := range matches {
		reqs, err := c.createRequestsForScheme(ctx, scheme, matched)
		if err != nil {
			return nil, err
		}
		requests = append(requests, reqs)
	}
	return requests, nil
}

// createRequestsForScheme is called for each unique scheme and the seeds that use it.
func (c crawler) createRequestsForScheme(ctx context.Context, scheme string, matches []cloudpath.Match) (download.Request, error) {
	switch scheme {
	case "s3":
		return c.s3Requests(ctx, matches)
	default:
		return nil, fmt.Errorf("unsupported scheme %q", scheme)
	}
}

func (c crawler) s3Requests(ctx context.Context, matched []cloudpath.Match) (download.Request, error) {
	if !c.Crawl.Download.AWS.AWS {
		return nil, fmt.Errorf("AWS authentication is required for S3")
	}
	awsConfig, err := awsconfig.LoadUsingFlags(ctx, c.Crawl.Download.AWS)
	if err != nil {
		return nil, err
	}
	container := s3fs.New(awsConfig)
	req := download.SimpleRequest{
		FS:   container,
		Mode: fs.FileMode(0600),
	}
	for _, match := range matched {
		req.Filenames = append(req.Filenames, match.Matched)
	}
	return req, nil
}
