// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package config

import "gopkg.in/yaml.v3"

type User struct {
	Email  string `yaml:"email"`
	UserID string `yaml:"user_id"`
	Name   string `yaml:"name"`
}

type Converter struct {
	// Name of the converter to use.
	ConverterName string `yaml:"converter"`

	ViewURLRewrites []string `yaml:"view_url_rewrites"` // Rewrite rules for viewurls specified as textutil.RewriteRules

	AllowAnonymousAccess bool `yaml:"allow_anonymous_access"` // allow anonymous access to the converted documents.

	// Default author to use if none can be obtained from the document itself.
	DefaultAuthor User `yaml:"default_author"`

	CustomConfig yaml.Node `yaml:"custom"`
}

type Converters []Converter