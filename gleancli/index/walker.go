// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package index

import (
	"context"
	"fmt"
	"path/filepath"

	"cloudeng.io/file/filewalk"
	"github.com/cosnicolaou/glean/gleancli/encoding"
)

type Walker struct {
	cnvs Converters
	wk   *filewalk.Walker
	ch   chan<- Request
}

func NewWalker(cnvs Converters, scanSize int, ch chan<- Request) *Walker {
	sc := filewalk.LocalFilesystem(scanSize)
	wk := filewalk.New(sc)
	return &Walker{
		wk:   wk,
		cnvs: cnvs,
		ch:   ch,
	}
}

func (w *Walker) dirs(ctx context.Context, prefix string, info *filewalk.Info, err error) (bool, []filewalk.Info, error) {
	return false, nil, nil
}

func (w *Walker) files(ctx context.Context, prefix string, info *filewalk.Info, ch <-chan filewalk.Contents) ([]filewalk.Info, error) {
	children := make([]filewalk.Info, 0, 10)
	var req Request
	for {
		var contents filewalk.Contents
		var ok bool
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case contents, ok = <-ch:
			if !ok {
				if len(req.Documents) > 0 {
					select {
					case <-ctx.Done():
						return nil, nil
					case w.ch <- req:
					}
				}
				return children, nil
			}
		}
		for _, file := range contents.Files {
			dl, err := encoding.ReadDownload(prefix, file.Name)
			if err != nil {
				fmt.Printf("failed to read download from %v: %v", filepath.Join(prefix, file.Name), err)
				continue
			}
			gd, err := w.cnvs.Convert(dl)
			if err != nil {
				fmt.Printf("failed to convert download from %v: to a Glean document: %v", filepath.Join(prefix, file.Name), err)
				continue
			}
			req.Documents = append(req.Documents, gd)
		}
		children = append(children, contents.Children...)
	}
}

func (w *Walker) Run(ctx context.Context, dir string) error {
	return w.wk.Walk(ctx, w.dirs, w.files, dir)
}
