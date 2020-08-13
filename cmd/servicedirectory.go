// Copyright © 2020 Cisco
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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/poller"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/queue"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/sdhandler"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/services"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var gcloudProject string
var gcloudRegion string
var datastore services.Datastore
var sendQueue queue.Queue
var sdHandler sdhandler.Handler

// servicedirectoryCmd represents the servicedirectory command
var servicedirectoryCmd = &cobra.Command{
	Use:   "servicedirectory",
	Short: "Connect to Service Directory to get registered services",
	Long: `This command connects to Google Cloud Service Directory and
observes changes in services published in it, i.e. metadata, addresses and
ports.

In order to work, a project, location and valid credentials must be provided.`,
	Run:     runServiceDirectory,
	Aliases: []string{"sd", "gcloud", "gcsd"},
}

func init() {
	rootCmd.AddCommand(servicedirectoryCmd)

	servicedirectoryCmd.Flags().StringVar(&gcloudProject, "project", "", "gcloud project name")
	servicedirectoryCmd.Flags().StringVar(&gcloudRegion, "region", "", "gcloud region location. Example: us-west2")
}

func runServiceDirectory(cmd *cobra.Command, args []string) {
	var err error
	l := log.With().Str("func", "cmd.runServiceDirectory").Logger()
	l.Info().Msg("starting...")

	ctx, canc := context.WithCancel(context.Background())

	// Parse flags
	if len(gcloudRegion) == 0 {
		l.Fatal().Err(fmt.Errorf("%s", "region not provided")).Msg("fatal error encountered")
	}
	if len(gcloudProject) == 0 {
		l.Fatal().Err(fmt.Errorf("%s", "project not provided")).Msg("fatal error encountered")
	}
	if len(credsPath) == 0 {
		l.Debug().Msg("resetting credentials path to ./credentials/gcloud-credentials.json")
		credsPath = "./credentials/gcloud-credentials.json"
	}

	// Get the handler
	sdHandler, err = sdhandler.New(ctx, gcloudRegion, metadataKey, gcloudProject, credsPath)
	if err != nil {
		l.Fatal().Err(err).Msg("error while trying to connect to service directory")
	}

	// Get the datastore
	datastore = services.NewDatastore()

	// Get the queue
	servsHandler, err := services.NewHandler(ctx, sanitizeAdaptorEndpoint(endpoint))
	if err != nil {
		l.Fatal().Err(err).Msg("error while trying to connect to service directory")
	}
	sendQueue = queue.New(ctx, servsHandler)

	// Get the poller
	poll := poller.New(ctx, interval)
	poll.SetPollFunction(processData)
	poll.Start()

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	<-sig
	fmt.Println()

	// Cancel the context and wait for objects that use it to receive
	// the stop command
	canc()

	l.Info().Msg("good bye!")
}

func processData() {
	data := sdHandler.GetServices()

	events := datastore.GetEvents(data)
	if len(events) > 0 {
		sendQueue.Enqueue(events)
	}
}
