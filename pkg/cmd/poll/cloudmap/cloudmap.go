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
	"strconv"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

const (
	awsIPv4Attr            string = "AWS_INSTANCE_IPV4"
	awsIPv6Attr            string = "AWS_INSTANCE_IPV6"
	awsPortAttr            string = "AWS_INSTANCE_PORT"
	awsDefaultInstancePort int32  = 80
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

func (a *awsCloudMap) getInstances(ctx context.Context, servID string) ([]*openapi.Service, error) {
	out, err := a.sd.ListInstancesWithContext(ctx, &servicediscovery.ListInstancesInput{ServiceId: &servID})
	if err != nil {
		return nil, err
	}

	oaSrvs := []*openapi.Service{}
	for _, inst := range out.Instances {
		oaSrv, err := a.parseInstance(servID, inst)
		if err != nil {
			log.Debug().Err(err).Str("service-id", servID).Msg("invalid instance: skipping...")
			continue
		}

		oaSrvs = append(oaSrvs, oaSrv)
	}

	return oaSrvs, nil
}

func (a *awsCloudMap) parseInstance(servID string, inst *servicediscovery.InstanceSummary) (*openapi.Service, error) {
	if inst.Id == nil || (inst.Id != nil && len(*inst.Id) == 0) {
		return nil, fmt.Errorf("found instance with no/empty ID")
	}

	// Check for metadata
	if inst.Attributes == nil {
		return nil, fmt.Errorf("instance doesn't have any attribute")
	}

	found := 0
	metadata := map[string]string{}
	for _, key := range a.targetKeys {
		if val, exists := inst.Attributes[key]; exists && val != nil && len(*val) > 0 {
			found++
			metadata[key] = *val
		}
	}
	if found != len(a.targetKeys) {
		return nil, fmt.Errorf("instance doesn't have required metadata keys")
	}

	// Check the address
	address := ""
	if ipv6 := inst.Attributes[awsIPv6Attr]; ipv6 != nil && len(*ipv6) > 0 {
		address = *ipv6
	}
	if ipv4 := inst.Attributes[awsIPv4Attr]; ipv4 != nil && len(*ipv4) > 0 {
		address = *ipv4
	}
	if len(address) == 0 {
		return nil, fmt.Errorf("instance has no address")
	}

	// Check the port
	port := awsDefaultInstancePort
	if instancePort := inst.Attributes[awsPortAttr]; instancePort != nil {
		if intPort, err := strconv.ParseInt(*instancePort, 10, 32); err == nil {
			port = int32(intPort)
		}
	}

	srv := &openapi.Service{
		Name:    *inst.Id,
		Address: address,
		Port:    port,
		Metadata: func() []openapi.Metadata {
			met := []openapi.Metadata{}
			for key, val := range metadata {
				met = append(met, openapi.Metadata{Key: key, Value: val})
			}
			return met
		}(),
	}

	return srv, nil
}
