// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config_test

import (
	"testing"

	"github.com/cosnicolaou/glean/gleancli/config"
)

const convertersSpec = `
- converter: "simpleHTML"
  default_author:
    email: "user@example.com"
  view_url_rewrites:
    - "s%a%b%"
  custom:
    foo: "bar"
    bar: "baz"
 `

func TestConverterCfg(t *testing.T) {
	converters, err := config.ParseConfig[config.Converters]([]byte(convertersSpec))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(converters), 1; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	cnv0 := converters[0]
	if got, want := cnv0.ConverterName, "simpleHTML"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	if got, want := cnv0.DefaultAuthor.Email, "user@example.com"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	custom := struct {
		Foo string `yaml:"foo"`
		Bar string `yaml:"bar"`
	}{}

	if err := cnv0.CustomConfig.Decode(&custom); err != nil {
		t.Fatal(err)
	}

	if got, want := custom.Foo, "bar"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := custom.Bar, "baz"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
