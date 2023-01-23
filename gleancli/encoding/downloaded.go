// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package encoding

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"

	"cloudeng.io/file/download"
)

func Marshal(v any) ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, v any) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(v)
	return err
}

func WriteDownload(dir, file string, v download.Result) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := Marshal(v)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, file), data, 0600)
}

func ReadDownload(dir, file string) (download.Result, error) {
	data, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		return download.Result{}, err
	}
	var v download.Result
	if err := Unmarshal(data, &v); err != nil {
		return download.Result{}, err
	}
	return v, nil
}
