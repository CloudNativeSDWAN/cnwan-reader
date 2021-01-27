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

package etcd

import (
	"context"
	"fmt"

	opsr "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry"
	opetcd "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry/etcd"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/queue"
	"go.etcd.io/etcd/clientv3"
	"gopkg.in/yaml.v3"
)

type etcdWatcher struct {
	options *Options
	keys    []string
	cli     *clientv3.Client
	kv      clientv3.KV
	watcher clientv3.Watcher
	queue.Queue
}

func (e *etcdWatcher) Watch(ctx context.Context) {
	log.Info().Msg("starting...")
	// TODO: implement me
}

func (e *etcdWatcher) getCurrentState(ctx context.Context, event string) (map[string]*openapi.Event, error) {
	resp, err := e.kv.Get(ctx, "namespaces", clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return nil, err
	}

	servs := map[string]*opsr.Service{}
	servsEndps := map[string][]*opsr.Endpoint{}

	for _, resp := range resp.Kvs {
		key := opetcd.KeyFromString(string(resp.Key))

		switch otype := key.ObjectType(); {
		case otype == opetcd.ServiceObject:
			var srv opsr.Service
			if err := yaml.Unmarshal(resp.Value, &srv); err != nil {
				log.Err(err).Str("key", key.String()).Msg("error while trying to unmarshal service, skipping...")
				continue
			}

			foundKeys := []string{}
			for _, keyToFind := range e.options.targetKeys {
				if _, exists := srv.Metadata[keyToFind]; exists {
					foundKeys = append(foundKeys, keyToFind)
				}
			}

			if len(foundKeys) == len(e.options.targetKeys) {
				servs[key.String()] = &srv
				servsEndps[key.String()] = []*opsr.Endpoint{}
			}

		case otype == opetcd.EndpointObject:
			var endp opsr.Endpoint
			if err := yaml.Unmarshal(resp.Value, &endp); err != nil {
				log.Err(err).Str("key", key.String()).Msg("error while trying to unmarshal endpoint, skipping...")
				continue
			}
			if len(endp.NsName) == 0 || len(endp.ServName) == 0 || len(endp.Name) == 0 {
				log.Error().Str("namespace", endp.NsName).Str("service", endp.ServName).
					Str("endpoint", endp.Name).Msg("endpoint is not valid as some names are unknown, skipping...")
				continue
			}

			srvKey := opetcd.KeyFromNames(endp.NsName, endp.ServName)
			servsEndps[srvKey.String()] = append(servsEndps[srvKey.String()], &endp)
		}
	}

	// Do the events
	events := map[string]*openapi.Event{}
	for srvKey, endpList := range servsEndps {
		srv, exists := servs[srvKey]
		if !exists {
			continue
		}

		for _, endp := range endpList {
			ev := openapi.Event{
				Event: event,
				Service: openapi.Service{
					Name:    endp.Name,
					Address: endp.Address,
					Port:    endp.Port,
				},
			}

			metadataList := []openapi.Metadata{}
			for key, val := range srv.Metadata {
				metadataList = append(metadataList, openapi.Metadata{Key: key, Value: val})
			}
			ev.Service.Metadata = metadataList

			evKey := fmt.Sprintf("%s:%d", endp.Address, endp.Port)
			events[evKey] = &ev
		}
	}

	return events, nil
}
