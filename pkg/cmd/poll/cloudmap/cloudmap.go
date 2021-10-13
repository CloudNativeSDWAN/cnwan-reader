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
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

const (
	awsIPv4Attr            string        = "AWS_INSTANCE_IPV4"
	awsIPv6Attr            string        = "AWS_INSTANCE_IPV6"
	awsPortAttr            string        = "AWS_INSTANCE_PORT"
	awsDefaultInstancePort int32         = 80
	defaultTimeout         time.Duration = 30 * time.Second
)

type awsCloudMap struct {
	opts *options
	sd   servicediscoveryiface.ServiceDiscoveryAPI
}

func (a *awsCloudMap) getServiceTags(ctx context.Context) (map[string]*openapi.Service, error) {
	out, err := a.sd.ListServicesWithContext(ctx, &servicediscovery.ListServicesInput{})
	if err != nil {
		return nil, err
	}

	keysMap := map[string]bool{}
	for _, key := range a.opts.keys {
		keysMap[key] = true
	}

	servTags := map[string]*openapi.Service{}
	for _, srv := range out.Services {
		l := log.With().Str("service-name", aws.StringValue(srv.Name)).Logger()

		metadata := func() map[string]string {
			tagsCtx, tagsCanc := context.WithTimeout(ctx, 30*time.Second)
			defer tagsCanc()

			out, err := a.sd.ListTagsForResourceWithContext(tagsCtx, &servicediscovery.ListTagsForResourceInput{
				ResourceARN: srv.Arn,
			})
			if err != nil {
				l.Warn().Err(err).Msg("could not get tags for service: skipping...")
				return map[string]string{}
			}

			tags := map[string]string{}
			for _, tag := range out.Tags {
				if _, exists := keysMap[aws.StringValue(tag.Key)]; exists {
					tags[aws.StringValue(tag.Key)] = aws.StringValue(tag.Value)
				}
			}
			return tags
		}()

		if len(metadata) == 0 {
			continue
		}

		endps, err := func() ([]*openapi.Service, error) {
			instCtx, instCanc := context.WithTimeout(ctx, 30*time.Second)
			defer instCanc()

			insts, err := a.sd.ListInstancesWithContext(instCtx, &servicediscovery.ListInstancesInput{
				ServiceId: srv.Id,
			})
			if err != nil {
				return nil, err
			}

			srvEp := []*openapi.Service{}
			for _, inst := range insts.Instances {
				srvEp = append(srvEp, &openapi.Service{
					Name:    aws.StringValue(inst.Id),
					Address: aws.StringValue(inst.Attributes["AWS_INSTANCE_IPV4"]),
					Port: func() int32 {
						val, _ := strconv.ParseInt(aws.StringValue(inst.Attributes["AWS_INSTANCE_PORT"]), 10, 32)
						return int32(val)
					}(),
				})
			}
			return srvEp, nil
		}()
		if err != nil {
			l.Err(err).Msg("error while getting instances for service: skipping...")
			continue
		}

		for _, endp := range endps {
			name := path.Join(aws.StringValue(srv.Name), endp.Name)
			servTags[name] = &openapi.Service{
				Name:    name,
				Address: endp.Address,
				Port:    endp.Port,
				Metadata: func() (met []openapi.Metadata) {
					for k, v := range metadata {
						met = append(met, openapi.Metadata{Key: k, Value: v})
					}
					return
				}(),
			}
		}
	}

	return servTags, nil
}

func (a *awsCloudMap) getCurrentState(ctx context.Context) (map[string]*openapi.Service, error) {
	srvCtx, srvCanc := context.WithTimeout(ctx, defaultTimeout)
	srvIDs, err := a.getServicesIDs(srvCtx)
	if err != nil {
		srvCanc()
		return nil, err
	}
	srvCanc()

	if len(srvIDs) == 0 {
		return map[string]*openapi.Service{}, nil
	}

	var wg sync.WaitGroup
	wg.Add(len(srvIDs))
	var locker sync.Mutex
	oaSrvs := map[string]*openapi.Service{}

	for _, srvID := range srvIDs {
		go func(id string) {
			defer wg.Done()
			instCtx, instCanc := context.WithTimeout(ctx, defaultTimeout)
			defer instCanc()

			insts, err := a.getInstances(instCtx, id)
			if err != nil {
				log.Err(err).Str("serv-id", id).Msg("could not get instances for this service, skipping...")
				return
			}

			locker.Lock()
			defer locker.Unlock()
			for i := 0; i < len(insts); i++ {
				oaID := fmt.Sprintf("services/%s/endpoints/%s", id, insts[i].Name)
				oaSrvs[oaID] = insts[i]
			}
		}(srvID)
	}
	wg.Wait()

	return oaSrvs, nil
}

func (a *awsCloudMap) getServicesIDs(ctx context.Context) ([]string, error) {
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
	for _, key := range a.opts.keys {
		if val, exists := inst.Attributes[key]; exists && val != nil && len(*val) > 0 {
			found++
			metadata[key] = *val
		}
	}
	if found != len(a.opts.keys) {
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
