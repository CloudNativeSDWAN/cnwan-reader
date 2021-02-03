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

	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

type awsCloudMap struct {
	opts *Options
	sd   servicediscoveryiface.ServiceDiscoveryAPI

	targetKeys      []string
	adaptorEndpoint string
}

func (a *awsCloudMap) getCurrentState(ctx context.Context) {
	// TODO: implement me
}

func (a *awsCloudMap) getServicesIDs(ctx context.Context) ([]string, error) {
	// TODO: specify that it takes only 100 services at a time
	// TODO: get next page?
	out, err := a.sd.ListServicesWithContext(ctx, &servicediscovery.ListServicesInput{})
	if err != nil {
		return nil, err
	}

	servIDs := []string{}
	for _, service := range out.Services {
		if service.Id != nil && len(*service.Id) > 0 {
			servIDs = append(servIDs, *service.Id)
		} else {
			log.Debug().Msg("found service with no/empty ID has been found: skipping...")
		}
	}

	return servIDs, nil
}
