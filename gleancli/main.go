// Copyright 2022 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"

	"cloudeng.io/cmdutil/subcmd"
	"github.com/cosnicolaou/glean/gleancli/config"
)

type GlobalFlags struct {
	config.ConfigFlags
}

var (
	globalFlags  GlobalFlags
	globalConfig *config.Config
	cmdSet       *subcmd.CommandSetYAML
)

func init() {
	cmdSet = subcmd.MustFromYAML(`name: gleancli
summary: command line for working with the Glean API
commands:
  - name: datasources
    summary: register or list Glean SDK data sources
    commands:
      - name: list
        summary: list the configuration for the specified data source
        arguments:
          - datasource
      - name: register
        summary: add the data source specified in a json file containing values
                 for gleansdk.CustomDatasourceConfig
        arguments:
          - datasource-config-file
`)

	cmdSet.Set("datasources", "list").RunnerAndFlags(
		listCmdRunner, subcmd.MustRegisteredFlagSet(&struct{}{}))

	cmdSet.Set("datasources", "register").RunnerAndFlags(
		registerCmdRunner, subcmd.MustRegisteredFlagSet(&struct{}{}))

	globals := subcmd.GlobalFlagSet()
	globals.MustRegisterFlagStruct(&globalFlags, nil, nil)
	cmdSet.WithGlobalFlags(globals)
	cmdSet.WithMain(mainWrapper)
}

func mainWrapper(ctx context.Context, cmdRunner func(ctx context.Context) error) error {
	cfg, err := config.ParseConfig(globalFlags.Config)
	if err != nil {
		return err
	}
	globalConfig = cfg
	return cmdRunner(ctx)
}

func main() {
	cmdSet.MustDispatch(context.Background())
}
