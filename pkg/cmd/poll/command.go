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

package poll

import (
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/cmd/poll/cloudmap"
	"github.com/spf13/cobra"
)

// GetPollCommand returns the poll command and all its subcommands
func GetPollCommand() *cobra.Command {
	// TODO: on next version this will probably be changed and adopt some
	// other programming pattern, maybe with a factory.

	cmd := &cobra.Command{
		TraverseChildren: true,
		Use:              pollUse,
		Short:            pollShort,
		Long:             pollLong,
		Example:          pollExample,
	}

	// Flags
	cmd.Flags().Int("poll-interval", 5, "interval between two consecutive polls")

	// Subcommands
	// TODO: service directory will be inserted here as well
	cmd.AddCommand(cloudmap.GetCloudMapCommand())

	return cmd
}
