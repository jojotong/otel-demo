/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"jojotong/otel-demo/pkg/client"
	"jojotong/otel-demo/pkg/observe"
	"jojotong/otel-demo/pkg/server"
	"jojotong/otel-demo/pkg/worker"

	"github.com/spf13/cobra"
)

var (
	// client
	serverAddr string
	interval   int
)

func main() {
	cmd := &cobra.Command{
		Use:   "oteldemo",
		Short: "jojotong otel demo",
	}
	clientcmd := &cobra.Command{
		Use:   "client",
		Short: "client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return client.Run(serverAddr, interval)
		},
	}
	clientcmd.Flags().StringVarP(&serverAddr, "server-addr", "s", "http://localhost:8080", "")
	clientcmd.Flags().IntVarP(&interval, "interval", "i", 10, "")

	servercmd := &cobra.Command{
		Use:   "server",
		Short: "api server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := observe.Init(context.TODO()); err != nil {
				return err
			}
			if err := server.Init(); err != nil {
				return err
			}
			return server.Run()
		},
	}
	servercmd.Flags().StringVarP(&server.Opts.WorkerAddr, "worker-addr", "w", "http://localhost:8081", "")
	servercmd.Flags().StringVarP(&server.Opts.MysqlAddr, "mysql-addr", "", "localhost:3306", "")
	servercmd.Flags().StringVarP(&server.Opts.MysqlRootPassword, "mysql-root-password", "", "", "")
	servercmd.Flags().StringVarP(&server.Opts.MysqlDBName, "mysql-db-name", "", "kubegems", "")

	workercmd := &cobra.Command{
		Use:   "worker",
		Short: "worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := observe.Init(context.TODO()); err != nil {
				return err
			}
			return worker.Run()
		},
	}
	cmd.AddCommand(clientcmd, servercmd, workercmd)
	if err := cmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
