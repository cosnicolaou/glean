// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cosnicolaou/gleansdk"
	"gopkg.in/yaml.v3"
)

type ConfigFlags struct {
	Config string `subcmd:"config,$HOME/.glean.yaml,'glean config file'"`
}

type Config struct {
	Auth struct {
		BearerToken string `yaml:"token"`
	}
	API struct {
		Domain string `yaml:"domain"`
	}
}

func (c *Config) String() string {
	var out strings.Builder
	out.WriteString("auth:\n")
	if len(c.Auth.BearerToken) > 0 {
		fmt.Fprintf(&out, "  token: **redacted**\n")
	}
	fmt.Fprintf(&out, "api\n  url: %s\n", c.API.Domain)
	return out.String()
}

func (c *Config) NewAPIClient(ctx context.Context) (context.Context, *gleansdk.APIClient) {
	templateVars := map[string]string{
		"domain": c.API.Domain,
	}
	ctx = context.WithValue(ctx, gleansdk.ContextAccessToken, c.Auth.BearerToken)
	ctx = context.WithValue(ctx, gleansdk.ContextServerVariables, templateVars)
	return ctx, gleansdk.NewAPIClient(gleansdk.NewConfiguration())
}

func ParseConfig(file string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("warning: %q not found\n", file)
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal([]byte(data), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
