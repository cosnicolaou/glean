// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package index

import (
	"fmt"
	"mime"
	"path"
	"strings"

	"cloudeng.io/file/download"
	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/gleansdk"
)

// Converter represents the ability to convert a download.Result into a
// gleansdk.DocumentDefinition for indexing.
type Converter interface {
	GleanDocument(dl download.Result) (*gleansdk.DocumentDefinition, error)
}

// Converters represents a list of converters applied in turn. The first
// one to return a non-nil document is used.
type Converters []Converter

// NewConverters creates instances of Converter for all of  the converters
// specified for the datasource.
func NewConverters(c config.Datasource) (Converters, error) {
	cnvs := []Converter{}
	for _, cnv := range c.Converters {
		c, err := NewConverter(c.CustomDatasourceConfig.GetName(), cnv)
		if err != nil {
			return nil, err
		}
		cnvs = append(cnvs, c)
	}
	return cnvs, nil
}

func (c Converters) Convert(dl download.Result) (*gleansdk.DocumentDefinition, error) {
	for _, cnv := range c {
		gd, err := cnv.GleanDocument(dl)
		if err != nil {
			return nil, err
		}
		if gd != nil {
			return gd, nil
		}
	}
	return nil, fmt.Errorf("no converter claimed %v", dl.Name)
}

func NewConverter(name string, c config.Converter) (Converter, error) {
	switch c.ConverterName {
	case "simpleHTML":
		return NewSimpleHTMLConverter(name, c)
	default:
		return nil, fmt.Errorf("unsupported converter: %v", c.ConverterName)
	}
}

func mimeTypeFromPath(p string) string {
	mimeType := mime.TypeByExtension(path.Ext(p))
	if idx := strings.Index(mimeType, ";"); idx > 0 {
		return mimeType[:idx]
	}
	return mimeType
}
