// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cosnicolaou/gleansdk"
	"gopkg.in/yaml.v3"
)

type ConfigFlags struct {
	Config string `subcmd:"config,$HOME/.glean.yaml,'glean config file'"`
}

type GleanConfig []struct {
	Name string `yaml:"name"`
	Auth struct {
		BearerToken string `yaml:"token"`
	}
	API struct {
		Domain string `yaml:"domain"`
	}
}

func (c GleanConfig) String() string {
	var out strings.Builder
	for _, cfg := range c {
		fmt.Fprintf(&out, "name: %s\n  auth:\n", cfg.Name)
		if len(cfg.Auth.BearerToken) > 0 {
			fmt.Fprintf(&out, "  token: **redacted**\n")
		}
		fmt.Fprintf(&out, "api\n  url: %s\n", cfg.API.Domain)
	}
	return out.String()
}

func (c GleanConfig) NewAPIClient(ctx context.Context, name string) (context.Context, *gleansdk.APIClient, error) {
	for _, cfg := range c {
		if cfg.Name == name {
			templateVars := map[string]string{
				"domain": cfg.API.Domain,
			}
			ctx = context.WithValue(ctx, gleansdk.ContextAccessToken, cfg.Auth.BearerToken)
			ctx = context.WithValue(ctx, gleansdk.ContextServerVariables, templateVars)
			return ctx, gleansdk.NewAPIClient(gleansdk.NewConfiguration()), nil
		}
	}
	return ctx, nil, fmt.Errorf("failed to find config for %s", name)
}

func ParseConfig[T any](buf []byte) (T, error) {
	var cfg T
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func ParseConfigFile[T any](file string) (T, error) {
	spec, err := os.ReadFile(file)
	if err != nil {
		var cfg T
		return cfg, err
	}
	return ParseConfig[T](spec)
}
