// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config_test

import (
	"reflect"
	"testing"

	"github.com/cosnicolaou/glean/gleancli/config"
)

const crawlsSpec = `
  name: test
  depth: 3
  seeds:
    - s3://foo/bar
    - https://yahoo.com

  download:
    default_concurrency: 4 # 0 will default to all available CPUs
    default_request_chan_size: 100
    default_downloads_chan_size: 100
    concurrency: [1, 2, 4]
`

func TestCrawlConfig(t *testing.T) {
	crawl, err := config.ParseConfig[config.Crawl]([]byte(crawlsSpec))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := crawl.Name, "test"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := len(crawl.Seeds), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := crawl.Depth, 3; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := crawl.Download.DefaultConcurrency, 4; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := crawl.Download.Concurrency, []int{1, 2, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

}
