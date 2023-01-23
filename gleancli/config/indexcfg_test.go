// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/cosnicolaou/glean/gleancli/config"
)

const indexSpec = `
  glean_instance: glean-dev
  force_restart: true
  force_deletion: true
  readdir_entries: 72
`

func TestIndexConfig(t *testing.T) {
	index, err := config.ParseConfig[config.Index]([]byte(indexSpec))
	if err != nil {
		t.Fatal(err)
	}

	if got, want := index.ForceDeletion, true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := index.ForceRestart, true; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := index.ReaddirEntries, 72; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
