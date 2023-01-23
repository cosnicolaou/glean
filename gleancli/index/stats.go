// Copyright 2023 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package index

import (
	"context"
	"fmt"

	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/gleansdk"
)

// StatsFlags represents the flags to the bulk indexing command.
type StatsFlags struct {
	config.FileFlags
}

func Stats(ctx context.Context, gleanConfig config.GleanConfig, fv *StatsFlags, datasource string) error {
	cfg, err := config.DatasourceForName(fv.ConfigFile, datasource)
	if err != nil {
		return err
	}
	ctx, client, err := gleanConfig.NewAPIClient(ctx, cfg.GleanInstance)
	if err != nil {
		return err
	}

	var req gleansdk.GetDocumentCountRequest
	req.SetName(cfg.CustomDatasourceConfig.GetName())

	count, resp, err := client.DocumentsApi.GetdocumentcountPost(ctx).GetDocumentCountRequest(req).Execute()
	if err := handleHTTPError(resp, err); err != nil {
		return err
	}

	fmt.Printf("Datasource: %v\n", req.GetName())
	fmt.Printf("\t# documents: %d\n", count.GetDocumentCount())

	return nil
}
