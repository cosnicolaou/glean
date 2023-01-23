// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"github.com/cosnicolaou/glean/gleancli/crawl"
)

type crawlCmd struct{}

func (cmd *crawlCmd) run(ctx context.Context, values interface{}, args []string) error {
	return crawl.Run(ctx, globalConfig, values.(*crawl.Flags), args[0])
}
