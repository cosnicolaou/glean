// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package index

import (
	"bytes"
	"fmt"

	"cloudeng.io/file/download"
	"cloudeng.io/text/textutil"
	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/gleansdk"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type SimpleHTMLConverter struct {
	name              string
	cfg               config.Converter
	view_url_rewrites textutil.RewriteRules
}

func NewSimpleHTMLConverter(name string, c config.Converter) (Converter, error) {
	rwr, err := textutil.NewRewriteRules(c.ViewURLRewrites...)
	if err != nil {
		return nil, err
	}
	return &SimpleHTMLConverter{
		name:              name,
		cfg:               c,
		view_url_rewrites: rwr,
	}, nil
}

func (cnv *SimpleHTMLConverter) GleanDocument(dl download.Result) (*gleansdk.DocumentDefinition, error) {
	mimeType := mimeTypeFromPath(dl.Name)
	if mimeType != "text/html" {
		return nil, fmt.Errorf("unsuported mime type: %v", mimeType)
	}
	gd := &gleansdk.DocumentDefinition{}
	gd.Datasource = cnv.name
	gd.SetId(dl.Name)
	gd.SetViewURL(cnv.view_url_rewrites.ReplaceAllStringFirst(dl.Name))

	var ht htmlTitle
	title := ht.htmlTitle(dl.Contents)
	gd.SetTitle(title)

	//	gd.Summary = &gleansdk.ContentDefinition{}
	//	gd.Summary.SetMimeType(mimeType)
	gd.Body = &gleansdk.ContentDefinition{}
	gd.Body.SetMimeType(mimeType)
	gd.Body.SetTextContent(string(dl.Contents))

	gd.Author = &gleansdk.UserReferenceDefinition{}
	gd.Author.SetEmail(cnv.cfg.DefaultAuthor.Email)

	gd.Permissions = &gleansdk.DocumentPermissionsDefinition{}
	gd.Permissions.SetAllowAnonymousAccess(true)

	gd.UpdatedAt = new(int64)
	*gd.UpdatedAt = int64(dl.FileInfo.ModTime().Unix())

	return gd, nil
}

type htmlTitle struct{}

func (cnv htmlTitle) htmlTitle(contents []byte) string {
	parsed, err := html.Parse(bytes.NewReader(contents))
	if err != nil {
		return ""
	}
	return cnv.htmlTitleTag(parsed)
}

func (cnv htmlTitle) htmlTitleTag(n *html.Node) string {
	if n.Type == html.ElementNode {
		if n.DataAtom == atom.Title {
			return n.FirstChild.Data
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := cnv.htmlTitleTag(c); len(title) > 0 {
			return title
		}
	}
	return ""
}
