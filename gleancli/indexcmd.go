// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"github.com/cosnicolaou/glean/gleancli/index"
)

type indexCmd struct{}

func (cmd *indexCmd) bulk(ctx context.Context, values interface{}, args []string) error {
	return index.Bulk(ctx, globalConfig, values.(*index.BulkFlags), args[0])
}

func (cmd *indexCmd) stats(ctx context.Context, values interface{}, args []string) error {
	return index.Stats(ctx, globalConfig, values.(*index.StatsFlags), args[0])
}
