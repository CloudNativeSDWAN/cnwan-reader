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

package cloudmap

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	log zerolog.Logger
)

func init() {
	output := zerolog.ConsoleWriter{Out: os.Stdout}
	log = zerolog.New(output).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// GetCloudMapCommand returns the cloudmap command
//
// TODO: on next version this will probably be changed and adopt some
// other programming pattern, maybe with a factory.
func GetCloudMapCommand() *cobra.Command {
	var cm *awsCloudMap

	cmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Long:    cmdLong,
		Example: cmdExample,
		PreRun: func(cmd *cobra.Command, _ []string) {
			// Parse the flag
			// TODO: implement me
		},
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: implement me
			ctx, canc := context.WithCancel(context.Background())
			exitChan := make(chan struct{})

			// TODO: use the context and the handler
			_, _ = ctx, cm

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
	cmd.Flags().String("region", "", "region to use")
	cmd.Flags().String("credentials-path", "", "the path to the credentials file")
	cmd.Flags().StringSlice("metadata-keys", []string{}, "the metadata keys to watch for")

	return cmd
}
