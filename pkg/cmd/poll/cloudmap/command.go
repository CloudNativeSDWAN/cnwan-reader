// Copyright © 2021 Cisco
//
// SPDX-License-Identifier: Apache-2.0
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

package cloudmap

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/configuration"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/poller"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/queue"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/services"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel)
}

// GetCloudMapCommand returns the cloudmap command
//
// TODO: on next version this will probably be changed and adopt some
// other programming pattern, maybe with a factory.
func GetCloudMapCommand() *cobra.Command {
	var cm *awsCloudMap
	var withTags bool

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
		PreRun: func(cmd *cobra.Command, _ []string) {
			opts, err := parseFlags(cmd, configuration.GetConfigFile())
			if err != nil {
				log.Fatal().Err(err).Msg("fatal error encountered")
				return
			}

			if len(opts.credsPath) > 0 {
				os.Setenv("AWS_SHARED_CREDENTIALS_FILE", opts.credsPath)
			}

			if opts.debug {
				log = log.Level(zerolog.DebugLevel)
			}

			sess, err := session.NewSession()
			if err != nil {
				log.Fatal().Err(err).Msg("could not start AWS session")
				return
			}
			sd := servicediscovery.New(sess, aws.NewConfig().WithRegion(opts.region))

			cm = &awsCloudMap{
				opts: opts,
				sd:   sd,
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			run(cm, withTags)
		},
	}

	// Flags
	cmd.Flags().String("region", "", "region to use")
	cmd.Flags().String("credentials-path", "", "the path to the credentials file")
	cmd.Flags().StringSlice("metadata-keys", []string{}, "the metadata keys to watch for")
	cmd.Flags().BoolVar(&withTags, "with-tags", false, "whether to look for AWS tags rather than attributes")

	return cmd
}

func run(cm *awsCloudMap, withTags bool) {
	log.Info().Str("service-registry", "Cloud Map").Str("adaptor", cm.opts.adaptor).Msg("starting...")
	if withTags {
		log.Info().Msg("switching to tag parsing...")
	}

	ctx, canc := context.WithCancel(context.Background())

	datastore := services.NewDatastore()
	servsHandler, err := services.NewHandler(ctx, cm.opts.adaptor)
	if err != nil {
		log.Fatal().Err(err).Msg("error while trying to connect to aws cloud map")
	}
	sendQueue := queue.New(ctx, servsHandler)

	go func() {
		log.Info().Msg("getting initial state...")
		var (
			oaSrvs map[string]*openapi.Service
			err    error
		)

		if !withTags {
			oaSrvs, err = cm.getCurrentState(ctx)
		} else {
			oaSrvs, err = cm.getServiceTags(ctx)
		}

		if err != nil {
			log.Fatal().Err(err).Msg("error while getting initial state of cloud map")
			return
		}

		log.Info().Msg("done")
		if filtered := datastore.GetEvents(oaSrvs); len(filtered) > 0 {
			go sendQueue.Enqueue(filtered)
		}

		// Get the poller
		log.Info().Msg("observing changes...")
		poll := poller.New(ctx, cm.opts.interval)
		poll.SetPollFunction(func() {
			var (
				oaSrvs map[string]*openapi.Service
				err    error
			)

			if !withTags {
				oaSrvs, err = cm.getCurrentState(ctx)
			} else {
				oaSrvs, err = cm.getServiceTags(ctx)
			}

			if err != nil {
				log.Err(err).Msg("error while polling, skipping...")
				return
			}

			if filtered := datastore.GetEvents(oaSrvs); len(filtered) > 0 {
				log.Info().Msg("changes detected")
				go sendQueue.Enqueue(filtered)
			}
		})

		poll.Start()
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

	log.Info().Msg("good bye!")
}
