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
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	l "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	debugMode   bool
	interval    int
	metadataKey string
	endpoint    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cnwan-reader",
	Short: "CN-WAN Reader observes changes in metadata in a service registry.",
	Long: `CN-WAN Reader connects to a service registry and 
observes changes about registered services, delivering found events to a
a separate handler for processing.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "whether to log debug lines")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", 5, "number of seconds between two consecutive polls")
	rootCmd.PersistentFlags().StringVar(&endpoint, "adaptor-api", "localhost/cnwan", "the api, in forrm of host:port/path, where the events will be sent to. Look at the documentation to learn more about this.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// -- Configure logger
	l.Logger = l.Output(zerolog.ConsoleWriter{
		Out: os.Stdout,
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("%s\t|", i)
		},
	}).Level(func() zerolog.Level {
		if debugMode {
			return zerolog.DebugLevel
		}

		return zerolog.InfoLevel
	}())
}

func sanitizeAdaptorEndpoint(endp string) string {
	endp = strings.Trim(endp, "/")

	if strings.HasPrefix(endp, "localhost") {
		// Replace localhost in case we are running insde docker
		if mode := os.Getenv("MODE"); len(mode) > 0 && mode == "docker" {
			return strings.Replace(endp, "localhost", "host.docker.internal", 1)
		}
	}

	return endp
}
