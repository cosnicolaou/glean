// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

import (
	"cloudeng.io/aws/awsconfig"
)

type Download struct {
	AWS                    awsconfig.AWSFlags `yaml:"aws,omitempty"`
	DefaultConcurrency     int                `yaml:"default_concurrency"`
	DefaultRequestChanSize int                `yaml:"default_request_chan_size"`
	DefaultCrawledChanSize int                `yaml:"default_crawled_chan_size"`
	Concurrency            []int              `yaml:"concurrency"`
	RequestChanSizes       []int              `yaml:"request_chan_sizes"`
	CrawledChanSizes       []int              `yaml:"crawled_chan_sizes"`
}

type Extractors struct {
	Concurrency int      `yaml:"concurrency"`
	Use         []string `yaml:"use,flow"`
}

type CrawlCache struct {
	CachePrefix           string `yaml:"cache_prefix"`
	CacheClearBeforeCrawl bool   `yaml:"cache_clear_before_crawl"`
}

type Crawl struct {
	Name          string
	CrawlCache    `yaml:",inline"`
	Depth         int
	Seeds         []string
	NoFollowRules []string `yaml:"nofollow"`
	FollowRules   []string `yaml:"follow"`
	RewriteRules  []string `yaml:"rewrite"`
	Download      Download
	Extractors    Extractors
}
