// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

import (
	"fmt"

	"github.com/cosnicolaou/gleansdk"
)

// FileFlags represents a command line flag for the datasource config file.
type FileFlags struct {
	ConfigFile string `subcmd:"datasource-configs,,datasource config file"`
}

// Datasource represents a single datasource or corpus to be crawled and
// indexed.
type Datasource struct {
	Datasource string      // Datasource name.
	Crawls     []Crawl     // Crawls that obtain data for this datasource.
	Index                  // Index configuration for this datasource.
	Cache                  // Cache configuration for this datasource.
	Converters []Converter // Converters (from download.Result to Glean document) configuration.

	// The Glean datasource configuration.
	DatasourceConfig `yaml:"datasource_config"`
}

// Cache represents a cache configuration.
type Cache struct {
	Path string // Path is the location of the cache for this datasource.
}

// DatasourceConfig represents the configuration of the datasource with
// Glean's API.
type DatasourceConfig struct {
	GleanInstance                   string `yaml:"glean_instance"`
	gleansdk.CustomDatasourceConfig `yaml:",inline"`
}

// Datasources represents a list of named datasources.
type Datasources []Datasource

// ConfigForName for returns the configuration for the named datasource.
func (d Datasources) ConfigForName(name string) (Datasource, bool) {
	for _, ds := range d {
		if ds.Datasource == name {
			return ds, true
		}
	}
	return Datasource{}, false
}

// DatasourceForName returns the datasource configuration for the named datasource
// read from the specified config file.
func DatasourceForName(filename string, name string) (Datasource, error) {
	if len(filename) == 0 {
		return Datasource{}, fmt.Errorf("no datasource config file specified")
	}
	cfg, err := ParseConfigFile[Datasources](filename)
	if err != nil {
		return Datasource{}, err
	}
	ds, ok := cfg.ConfigForName(name)
	if !ok {
		return Datasource{}, fmt.Errorf("no datasource config found for %q in %q", name, filename)
	}
	return ds, nil
}
