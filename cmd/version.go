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

	"github.com/spf13/cobra"
)

var shortVer bool

const (
	longVersion  = "MAJOR=1; MINOR=2; GIT-VERSION=v.1.2.0"
	shortVersion = "v1.2.0"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the running version",
	Long: `Version prints the version of the running instance of CN-WAN Reader.
	
The message displayed will contain information about the MAJOR version, the
MINOR version and the git version.

When run with --short a shorter and human-readable version will be printed
instead`,
	Run: func(cmd *cobra.Command, args []string) {
		if shortVer {
			fmt.Println(shortVersion)
			return
		}

		fmt.Println(longVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&shortVer, "short", "s", false, "print a short version")
}
