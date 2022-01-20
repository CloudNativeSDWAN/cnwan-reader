// Copyright © 2020 Cisco
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

package sdhandler

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"

	sd "cloud.google.com/go/servicedirectory/apiv1beta1"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	sdpb "google.golang.org/genproto/googleapis/cloud/servicedirectory/v1beta1"
)

type gcloudServDir struct {
	metadataKey string
	region      string
	project     string
	ctx         context.Context
	cl          *sd.RegistrationClient
	baseParent  string
}

// New returns a handler for gcloud service directory
func New(ctx context.Context, region, metadataKey, project, credsPath string) (Handler, error) {
	jsonBytes, err := ioutil.ReadFile(credsPath)
	if err != nil {
		return nil, err
	}

	c, err := sd.NewRegistrationClient(ctx, option.WithCredentialsJSON(jsonBytes))
	if err != nil {
		return nil, err
	}

	return &gcloudServDir{
		region:      region,
		project:     project,
		metadataKey: metadataKey,
		ctx:         ctx,
		cl:          c,
		baseParent:  path.Join("projects", project, "locations", region),
	}, nil
}

// GetServices loads data from the service
func (g *gcloudServDir) GetServices() map[string]*openapi.Service {
	l := log.With().Str("func", "Handler.GetServices").Logger()
	maps := map[string]*openapi.Service{}

	nsList, err := g.getNamespacesList()
	if err != nil {
		l.Error().Err(err).Msg("error while getting namespaces list")
	}

	for _, ns := range nsList {
		l := l.With().Str("ns-name", ns.Name).Logger()

		servList, err := g.getServicesList(ns.Name)
		if err != nil {
			l.Warn().Err(err).Msg("error while getting services")
			continue
		}

		for _, serv := range servList {
			l := l.With().Str("service-name", serv.Name).Logger()

			epList, err := g.getEndpointsList(serv.Name)
			if err != nil {
				l.Warn().Err(err).Msg("error while getting endpoints")
				continue
			}

			for _, endpoint := range epList {
				l := l.With().Str("endpoint-name", endpoint.Name).Str("endpoint-address", endpoint.Address).
					Int32("endpoint-port", endpoint.Port).Logger()

				data := g.formatData(endpoint, serv.Metadata)

				if data != nil {
					l.Debug().Msg("endpoint has the required metadata key")
					mapKey := fmt.Sprintf("%s_%d", data.Address, data.Port)
					maps[mapKey] = data
				}
			}
		}
	}

	return maps
}

func (g *gcloudServDir) getNamespacesList() ([]*sdpb.Namespace, error) {
	req := &sdpb.ListNamespacesRequest{
		Parent: g.baseParent,
	}
	nsList := []*sdpb.Namespace{}

	// -- Get the list
	it := g.cl.ListNamespaces(g.ctx, req)
	if it == nil {
		return nsList, nil
	}

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		nsList = append(nsList, resp)
	}

	return nsList, nil
}

func (g *gcloudServDir) getServicesList(nsName string) ([]*sdpb.Service, error) {
	req := &sdpb.ListServicesRequest{
		Parent: nsName,
	}
	servList := []*sdpb.Service{}

	// -- Get the list
	it := g.cl.ListServices(g.ctx, req)
	if it == nil {
		return servList, nil
	}

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		servList = append(servList, resp)
	}

	return servList, nil
}

func (g *gcloudServDir) getEndpointsList(serv string) ([]*sdpb.Endpoint, error) {
	req := &sdpb.ListEndpointsRequest{
		Parent: serv,
	}
	endpointsList := []*sdpb.Endpoint{}

	// -- Get the list
	it := g.cl.ListEndpoints(g.ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		endpointsList = append(endpointsList, resp)
	}

	return endpointsList, nil
}

func (g *gcloudServDir) formatData(endpoint *sdpb.Endpoint, serviceMetadata map[string]string) *openapi.Service {
	metadataValue, exists := serviceMetadata[g.metadataKey]
	if !exists {
		return nil
	}

	if len(endpoint.Address) == 0 {
		return nil
	}

	return &openapi.Service{
		Address:  endpoint.Address,
		Name:     endpoint.Name,
		Metadata: []openapi.Metadata{{Key: g.metadataKey, Value: metadataValue}},
		Port:     endpoint.Port,
	}
}
