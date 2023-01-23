// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package crawl

import (
	"context"
	"fmt"

	"cloudeng.io/file/download"
	"github.com/cosnicolaou/glean/gleancli/config"
)

type DownloaderFactory struct {
	config.Download
	ProgressChan chan download.Progress
}

func depthOrDefault(depth int, values []int, def int) int {
	if depth < len(values) {
		return values[depth]
	}
	return def
}

func (df *DownloaderFactory) New(ctx context.Context, depth int) (
	downloader download.T,
	inputCh chan download.Request,
	outputCh chan download.Downloaded) {
	concurrency := depthOrDefault(depth, df.Concurrency, df.DefaultConcurrency)
	reqChanSize := depthOrDefault(depth, df.RequestChanSizes, df.DefaultRequestChanSize)
	dlChanSize := depthOrDefault(depth, df.CrawledChanSizes, df.DefaultCrawledChanSize)
	inputCh = make(chan download.Request, reqChanSize)
	outputCh = make(chan download.Downloaded, dlChanSize)
	downloader = download.New(download.WithNumDownloaders(concurrency))
	return
}

func (df *DownloaderFactory) HandleProgress(ctx context.Context, name string, progress <-chan download.Progress) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-progress:
			fmt.Printf(" 16%v: % 8v: % 8v\n", name, p.Downloaded, p.Outstanding)
		}
	}
}
