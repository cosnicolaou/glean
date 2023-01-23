// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosnicolaou/glean/gleancli/config"
	"github.com/cosnicolaou/gleansdk"
	"gopkg.in/yaml.v3"
)

type DatasourceConfigFile struct {
	ConfigFile string `subcmd:"datasource-configs,,datasource config file"`
}

type DatasourceFlags struct {
	DatasourceConfigFile
}

type datasourceCmds struct{}

func (ds datasourceCmds) download(ctx context.Context, values interface{}, args []string) error {
	instance := args[0]
	ctx, client, err := globalConfig.NewAPIClient(ctx, instance)
	if err != nil {
		return err
	}
	return downloadDataSource(ctx, client, args[1])
}

func (ds datasourceCmds) register(ctx context.Context, values interface{}, args []string) error {
	datasource := args[0]
	fv := values.(*DatasourceFlags)
	cfg, err := config.DatasourceForName(fv.ConfigFile, datasource)
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(cfg.CustomDatasourceConfig)
	if err != nil {
		return err
	}
	fmt.Printf("Registering custom datasource:\n%s\n", buf)

	ctx, client, err := globalConfig.NewAPIClient(ctx, cfg.GleanInstance)
	if err != nil {
		return err
	}
	getDatasourceConfigRequest := gleansdk.NewGetDatasourceConfigRequest()
	getDatasourceConfigRequest.Name = &datasource
	r, err := client.DatasourcesApi.AdddatasourcePost(ctx).CustomDatasourceConfig(cfg.CustomDatasourceConfig).Execute()
	return parseError(r, err)
}

func (ds datasourceCmds) config(ctx context.Context, values interface{}, args []string) error {
	cfg, err := config.ParseConfigFile[config.Datasources](args[0])
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", buf)
	return nil
}

func downloadDataSource(ctx context.Context, client *gleansdk.APIClient, datasource string) error {
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
