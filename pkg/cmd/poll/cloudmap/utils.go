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
	"fmt"

	"github.com/spf13/cobra"
)

func parseFlags(cmd *cobra.Command) (*Options, error) {
	opts := &Options{}

	awsRegion, _ := cmd.Flags().GetString("region")
	if len(awsRegion) == 0 {
		return nil, fmt.Errorf("region not provided")
	}
	opts.Region = awsRegion

	credsPath, _ := cmd.Flags().GetString("credentials-path")
	if len(credsPath) > 0 {
		opts.CredentialsPath = credsPath
	}

	pollInterval, _ := cmd.Flags().GetInt("poll-interval")
	if pollInterval <= 0 {
		pollInterval = 5
	}
	opts.PollInterval = pollInterval

	return opts, nil
}
