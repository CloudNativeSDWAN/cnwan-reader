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

package watch

import (
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/cmd/watch/etcd"
	"github.com/spf13/cobra"
)

// GetWatchCommand returns the watch commands and all its subcommands
func GetWatchCommand() *cobra.Command {
	// TODO: on next version this will probably be changed and adopt some
	// other programming pattern, maybe with a factory.

	cmd := &cobra.Command{
		Use:   "watch [COMMAND] [flags]",
		Short: "watch for changes",
		Long: `watch uses a watching mechanism to get changes in the
service registry. Not all service registries work this way: include --help or
-h to know which ones are included.`,
		Example: "watch etcd --endpoints localhost:2379 --username user --password pass",
	}

	// Subcommands
	cmd.AddCommand(etcd.GetEtcdCommand())

	return cmd
}
