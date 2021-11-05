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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	opetcd "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry/etcd"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/queue"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/services"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	namespace "go.etcd.io/etcd/client/v3/namespace"
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

			sr := opetcd.NewServiceRegistryWithEtcd(context.Background(), cli, &options.Prefix)

			watcher = &etcdWatcher{
				options: options,
				cli:     cli,
				kv:      namespace.NewKV(cli.KV, options.Prefix),
				watcher: namespace.NewWatcher(cli.Watcher, options.Prefix),
				servreg: sr,
			}
		},
		Run: func(cmd *cobra.Command, args []string) {

			defer watcher.cli.Close()

			// Get the adaptor endpoint
			fromFlag, _ := cmd.Flags().GetString("adaptor-api")
			adaptorEndpoint, err := sanitizeLocalhost(fromFlag, os.Getenv("MODE"))
			if err != nil {
				log.Err(err).Str("adaptor-endpoint", fromFlag).Msg("adaptor endpoint doesn't seem valid")
				return
			}

			// Get create events
			log.Info().Msg("getting current state of service registry from etcd...")
			currStateCtx, currStateCanc := context.WithTimeout(context.Background(), time.Minute)
			initialEvents, err := watcher.getCurrentState(currStateCtx, "create")
			currStateCanc()
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					log.Err(err).Int("seconds", 60).Msg("timeout expired while getting current state (did you specify the correct --endpoints ?)")
					return
				}

				log.Err(err).Msg("error while retrieving current state from etcd")
				return
			}

			ctx, canc := context.WithCancel(context.Background())
			exitChan := make(chan bool)

			// Get the queue and send the events
			servsHandler, err := services.NewHandler(ctx, adaptorEndpoint)
			if err != nil {
				log.Err(err).Msg("error while trying to connect to service directory")
				canc()
				return
			}
			watcher.Queue = queue.New(ctx, servsHandler)
			if len(initialEvents) > 0 {
				go watcher.Enqueue(initialEvents)
			}

			go func() {
				log.Info().Msg("watching for changes...")
				watcher.Watch(ctx)
				close(exitChan)
			}()

			// Graceful shutdown
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt)

			<-sig
			fmt.Println()
			log.Info().Msg("exit requested")

			// Cancel the context and wait for objects that use it to receive
			// the stop command
			canc()
			<-exitChan

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
