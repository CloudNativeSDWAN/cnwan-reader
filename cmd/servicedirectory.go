// Copyright © 2020, 2021 Cisco
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

var (
	gcloudProject     string
	gcloudRegion      string
	gcloudServAccount string
	datastore         services.Datastore
	sendQueue         queue.Queue
	sdHandler         sdhandler.Handler
)

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
	servicedirectoryCmd.Flags().StringVar(&gcloudServAccount, "service-account", "", "path to the gcloud service account. Example: ./service-account.json")
	servicedirectoryCmd.Flags().StringVar(&metadataKey, "metadata-key", "", "name of the metadata key to look for")
}

func parseServiceDirectoryConf(conf *ServiceDirectoryConfig) {
	if len(gcloudProject) == 0 && len(conf.ProjectID) > 0 {
		gcloudProject = conf.ProjectID
	}

	if len(gcloudRegion) == 0 && len(conf.Region) > 0 {
		gcloudRegion = conf.Region
	}

	if len(gcloudServAccount) == 0 && len(conf.ServiceAccountPath) > 0 {
		gcloudServAccount = conf.ServiceAccountPath
	}

	if interval == 0 && conf.PollingInterval > 0 {
		interval = conf.PollingInterval
	}

	if interval <= 0 {
		logger.Warn().Msg("invalid interval value used, using default...")
		interval = 5
	}
}

func validateSDFlags() error {
	if len(metadataKey) == 0 {
		return fmt.Errorf("error: no metadata key set")
	}

	if len(gcloudProject) == 0 {
		return fmt.Errorf("error: no gcloud project name set")
	}

	if len(gcloudRegion) == 0 {
		return fmt.Errorf("error: no gcloud region set")
	}

	if len(gcloudServAccount) == 0 {
		return fmt.Errorf("error: no service account path set")
	}

	return nil
}

func runServiceDirectory(cmd *cobra.Command, args []string) {
	var err error
	l := log.With().Str("func", "cmd.runServiceDirectory").Logger()

	if err := validateSDFlags(); err != nil {
		cmd.Usage()
		logger.Fatal().Err(err).Msg("error while starting service directory")
		os.Exit(1)
	}

	l.Info().Msg("starting...")

	ctx, canc := context.WithCancel(context.Background())

	// Get the handler
	sdHandler, err = sdhandler.New(ctx, gcloudRegion, metadataKey, gcloudProject, gcloudServAccount)
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
		go sendQueue.Enqueue(events)
	}
}
