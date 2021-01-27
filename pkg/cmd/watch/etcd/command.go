// Copyright © 2021 Cisco
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// All rights reserved.

package etcd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/namespace"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// GetEtcdCommand returns the etcd command
//
// TODO: on next version this will probably be changed and adopt some
// other programming pattern, maybe with a factory.
func GetEtcdCommand() *cobra.Command {
	var watcher *etcdWatcher

	cmd := &cobra.Command{
		Use:     etcdUse,
		Short:   etcdShort,
		Long:    etcdLong,
		Example: etcdExample,
		PreRun: func(cmd *cobra.Command, _ []string) {
			// Parse the flags
			options, err := parseFlags(cmd)
			if err != nil {
				log.Fatal().Err(err).Msg("error while parsing commands, check usage with --help")
				return
			}

			// Get the etcd clients
			cli, err := clientv3.New(getEtcdClientConfig(options))
			if err != nil {
				log.Fatal().Err(err).Msg("error while establishing connection to etcd client")
				return
			}

			watcher = &etcdWatcher{
				options: options,
				cli:     cli,
				kv:      namespace.NewKV(cli.KV, options.Prefix),
				watcher: namespace.NewWatcher(cli.Watcher, options.Prefix),
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			ctx, canc := context.WithCancel(context.Background())

			// TODO: do something with the client
			_, _ = watcher, ctx

			// Graceful shutdown
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)

			<-sig
			fmt.Println()
			log.Info().Msg("exit requested")

			// Cancel the context and wait for objects that use it to receive
			// the stop command
			canc()

			log.Info().Msg("good bye!")
		},
	}

	// Flags
	cmd.Flags().StringSlice("endpoints", func() []string {
		sanitized, _ := sanitizeLocalhost(defaultHost, os.Getenv("MODE"))
		return []string{fmt.Sprintf("%s:%d", sanitized, defaultPort)}
	}(), "endpoints where to connect to")
	cmd.Flags().String("username", "", "the username to authenticate as")
	cmd.Flags().String("password", "", "the password to use for this user")
	cmd.Flags().String("prefix", "/", "the prefix to include for all objects")
	cmd.Flags().StringSlice("metadata-keys", []string{}, "the metadata keys to look for")

	return cmd
}
