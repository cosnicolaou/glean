// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cosnicolaou/gleansdk"
)

func listCmdRunner(ctx context.Context, values interface{}, args []string) error {
	ctx, client := globalConfig.NewAPIClient(ctx)
	return listDataSource(ctx, client, args[0])
}

func registerCmdRunner(ctx context.Context, values interface{}, args []string) error {
	ctx, client := globalConfig.NewAPIClient(ctx)
	return registerDataSource(ctx, client, args[0])
}

func parseDatasourceConfig(file string) (*gleansdk.CustomDatasourceConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	cfg := &gleansdk.CustomDatasourceConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, err
}

func registerDataSource(ctx context.Context, client *gleansdk.APIClient, datasource string) error {
	dsConfig, err := parseDatasourceConfig(datasource)
	if err != nil {
		return fmt.Errorf("failed to find or parse datasource config from file: %q: %v", datasource, err)
	}
	getDatasourceConfigRequest := gleansdk.NewGetDatasourceConfigRequest()
	getDatasourceConfigRequest.Name = &datasource
	r, err := client.DatasourcesApi.AdddatasourcePost(ctx).CustomDatasourceConfig(*dsConfig).Execute()
	return parseError(r, err)
}

func listDataSource(ctx context.Context, client *gleansdk.APIClient, datasource string) error {
	getDatasourceConfigRequest := gleansdk.NewGetDatasourceConfigRequest()
	getDatasourceConfigRequest.Name = &datasource
	resp, r, err := client.DatasourcesApi.GetdatasourceconfigPost(ctx).GetDatasourceConfigRequest(*getDatasourceConfigRequest).Execute()
	if err != nil {
		return parseError(r, err)
	}
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
