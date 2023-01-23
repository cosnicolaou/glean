// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/cosnicolaou/glean/gleancli/config"
)

const configSpec = `
- name: glean-dev
  auth:
    token: "token"
  api:
    domain: glean-dev
- name: another-instance
  auth:
    token: "another token"
  api:
    domain: another-instance.com
`

func TestConfig(t *testing.T) {
	cfg, err := config.ParseConfig[config.GleanConfig]([]byte(configSpec))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(cfg), 2; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := cfg[1].Name, "another-instance"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
