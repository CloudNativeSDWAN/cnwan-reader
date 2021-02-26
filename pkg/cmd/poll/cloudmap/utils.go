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

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/configuration"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/internal/utils"
	"github.com/spf13/cobra"
)

func parseFlags(cmd *cobra.Command, conf *configuration.Config) (*options, error) {
	opts := &options{}

	if conf == nil {
		conf = &configuration.Config{
			ServiceRegistry: &configuration.ServiceRegistrySettings{
				AWSCloudMap: &configuration.CloudMapConfig{},
			},
		}
	}
	cmConf := conf.ServiceRegistry.AWSCloudMap

	awsRegion, _ := cmd.Flags().GetString("region")
	if len(awsRegion) == 0 {
		if len(cmConf.Region) == 0 {
			return nil, fmt.Errorf("region not provided")
		}

		awsRegion = cmConf.Region
	}
	opts.region = awsRegion

	credsPath, _ := cmd.Flags().GetString("credentials-path")
	if len(credsPath) == 0 {
		if len(cmConf.CredentialsPath) > 0 {
			credsPath = cmConf.CredentialsPath
		}
	}
	opts.credsPath = credsPath

	pollInterval := 5
	if cmd.Flags().Changed("poll-interval") {
		_pollInterval, _ := cmd.Flags().GetInt("poll-interval")
		if _pollInterval > 0 {
			pollInterval = _pollInterval
		}
	} else {
		if cmConf.PollInterval > 0 {
			pollInterval = cmConf.PollInterval
		}
	}
	opts.interval = pollInterval

	keys, err := utils.GetMetadataKeysFromCmdFlags(cmd)
	if err != nil {
		return nil, err
	}
	opts.keys = keys

	adaptor, err := utils.GetAdaptorEndpointFromFlags(cmd)
	if err != nil {
		return nil, err
	}
	opts.adaptor = adaptor
	opts.debug = utils.GetDebugModeFromFlags(cmd)

	return opts, nil
}
