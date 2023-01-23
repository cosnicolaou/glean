// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

type Index struct {
	ForceDeletion  bool `yaml:"force_deletion"`  // Glean's force deletion flag
	ForceRestart   bool `yaml:"force_restart"`   // Glean's force restart flag
	ReaddirEntries int  `yaml:"readdir_entries"` // number of entries per Readdir call.
}
