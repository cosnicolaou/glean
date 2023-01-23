// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cosnicolaou/gleansdk"
)

func parseError(r *http.Response, err error) error {
	if err == nil {
		return nil
	}
	oapiErr, ok := err.(*gleansdk.GenericOpenAPIError)
	if !ok {
		if r != nil {
			return fmt.Errorf("%v: %v: %v\n", r.Request.URL, r.StatusCode, err)
		}
		return err
	}
	if r == nil {
		return err
	}
	var tmp any
	body := oapiErr.Body()
	if json.Unmarshal(body, &tmp) == nil {
		return fmt.Errorf("%s: %v", body, err)
	}
	if body, nerr := io.ReadAll(r.Body); nerr == nil {
		return fmt.Errorf("%s: %v", body, err)
	}
	return err
}
