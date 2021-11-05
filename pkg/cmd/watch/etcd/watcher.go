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
	"github.com/google/go-cmp/cmp"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

type etcdWatcher struct {
	options *Options
	cli     *clientv3.Client
	kv      clientv3.KV
	watcher clientv3.Watcher
	queue.Queue
	servreg opsr.ServiceRegistry
}

func (e *etcdWatcher) Watch(ctx context.Context) {
	log.Info().Msg(e.options.Prefix)
	wchan := e.watcher.Watch(ctx, "", clientv3.WithPrefix(), clientv3.WithPrevKV())
	defer e.watcher.Close()

	for wresp := range wchan {
		for _, ev := range wresp.Events {

			key := opetcd.KeyFromString(string(ev.Kv.Key))
			var eventsToSend map[string]*openapi.Event

			switch evType := ev.Type; {
			case evType == mvccpb.DELETE:
				if key.ObjectType() == opetcd.EndpointObject && ev.PrevKv != nil && ev.PrevKv.Value != nil {
					log.Info().Str("key", key.String()).Msg("detected deleted endpoint")
					if endpEv, err := e.parseEndpointAndCreateEvent(ev.PrevKv, "delete"); err == nil && endpEv != nil {
						eventsToSend = map[string]*openapi.Event{key.String(): endpEv}
					}
				}
			case evType == mvccpb.PUT && ev.IsCreate():
				if key.ObjectType() == opetcd.EndpointObject && ev.Kv.Value != nil {
					log.Info().Str("key", key.String()).Msg("new endpoint detected")
					if endpEv, err := e.parseEndpointAndCreateEvent(ev.PrevKv, "create"); err == nil && endpEv != nil {
						eventsToSend = map[string]*openapi.Event{key.String(): endpEv}
					}
				}
			case evType == mvccpb.PUT && ev.IsModify():
				if key.ObjectType() == opetcd.EndpointObject {
					log.Info().Str("key", key.String()).Msg("detected updated endpoint")
					if endpEv, err := e.parseEndpointChange(ev.Kv, ev.PrevKv); err == nil && endpEv != nil {
						eventsToSend = map[string]*openapi.Event{key.String(): endpEv}
					}
				}
				if key.ObjectType() == opetcd.ServiceObject {
					log.Info().Str("key", key.String()).Msg("detected updated service")
					if endpEv, err := e.parseServiceChange(ev.Kv, ev.PrevKv); err == nil && endpEv != nil {
						eventsToSend = endpEv
					}
				}
			}

			if e.Queue != nil && len(eventsToSend) > 0 {
				go e.Queue.Enqueue(eventsToSend)
			}
		}
	}
}

func (e *etcdWatcher) parseEndpointAndCreateEvent(kvpair *mvccpb.KeyValue, eventName string) (*openapi.Event, error) {
	key := opetcd.KeyFromString(string(kvpair.Key))
	l := log.With().Str("key", key.String()).Str("event", eventName).Logger()

	endp, err := validateEndpointFromEtcd(kvpair.Value)
	if err != nil {
		l.Err(err).Msg("endpoint is not valid: skipping...")
		return nil, err
	}

	srv, err := e.servreg.GetServ(endp.NsName, endp.ServName)
	if err != nil {
		l.Err(err).Msg("error while trying to get parent service: skipping endpoint...")
		return nil, err
	}

	if len(e.options.targetKeys) > 0 && !mapContainsKeys(srv.Metadata, e.options.targetKeys) {
		l.Info().Msg("endpoint's parent service doesn't have target metadata keys: skipping...")
		return nil, nil
	}

	event := createOpenapiEvent(endp, srv, eventName)
	return event, nil
}

func (e *etcdWatcher) parseEndpointChange(now, prev *mvccpb.KeyValue) (*openapi.Event, error) {
	l := log.With().Str("key", string(now.Key)).Str("event", "update").Logger()
	var parsedPrev *opsr.Endpoint
	var parsedNow *opsr.Endpoint
	var nowErr error

	if now.Value != nil {
		_parsedNow, err := validateEndpointFromEtcd(now.Value)
		if err == nil {
			parsedNow = _parsedNow
		} else {
			nowErr = err
		}
	}
	if prev.Value != nil {
		_parsedPrev, err := validateEndpointFromEtcd(prev.Value)
		if err == nil {
			parsedPrev = _parsedPrev
		}
	}

	if parsedNow == nil && parsedPrev == nil {
		// Still invalid
		l.Info().Err(nowErr).Msg("endpoint is still not valid, skipping...")
		return nil, nil
	}

	// No need to check for error from keybuilder: if you're here it means
	// that they are indeed valid.
	var keyBuilder *opetcd.KeyBuilder
	if parsedNow != nil {
		keyBuilder, _ = opetcd.KeyFromServiceRegistryObject(parsedNow)
	} else {
		keyBuilder, _ = opetcd.KeyFromServiceRegistryObject(parsedPrev)
	}

	srv, err := e.servreg.GetServ(keyBuilder.GetNamespace(), keyBuilder.GetService())
	if err != nil {
		l.Err(err).Msg("error while retrieving parent service: skipping...")
		return nil, err
	}

	if !mapContainsKeys(srv.Metadata, e.options.targetKeys) {
		l.Info().Msg("endpoint's parent service doesn't have target metadata keys: skipping...")
		return nil, nil
	}

	// TODO: on future versions, this will be removed, in favor of a
	// simple map[string]string, the ones used by the operator
	parsedMetadata := []openapi.Metadata{}
	for key, val := range srv.Metadata {
		parsedMetadata = append(parsedMetadata, openapi.Metadata{Key: key, Value: val})
	}

	// It is not valid now
	if parsedNow == nil {
		l.Warn().Err(nowErr).Msg("endpoint seems to be not valid anymore and must be deleted")
		return &openapi.Event{
			Event: "delete",
			Service: openapi.Service{
				Name:     parsedPrev.Name,
				Address:  parsedPrev.Address,
				Port:     parsedPrev.Port,
				Metadata: parsedMetadata,
			},
		}, nil
	}

	// It was not valid before
	if parsedPrev == nil {
		l.Info().Msg("endpoint is now valid")
		return &openapi.Event{
			Event: "create",
			Service: openapi.Service{
				Name:     parsedNow.Name,
				Address:  parsedNow.Address,
				Port:     parsedNow.Port,
				Metadata: parsedMetadata,
			},
		}, nil
	}

	// What changed?
	parsedNow.Metadata = map[string]string{}
	parsedPrev.Metadata = map[string]string{}
	if !cmp.Equal(parsedNow, parsedPrev) {
		l.Info().Msg("endpoint effectively changed")
		return &openapi.Event{
			Event: "update",
			Service: openapi.Service{
				Name:     parsedNow.Name,
				Address:  parsedNow.Address,
				Port:     parsedNow.Port,
				Metadata: parsedMetadata,
			},
		}, nil
	}

	l.Info().Msg("no relevant changes detected: skipping...")
	return nil, nil
}

func (e *etcdWatcher) parseServiceChange(now, prev *mvccpb.KeyValue) (map[string]*openapi.Event, error) {
	l := log.With().Str("key", string(now.Key)).Str("event", "update").Logger()
	var parsedPrev *opsr.Service
	var parsedNow *opsr.Service

	if now.Value != nil {
		_parsedNow, err := validateServiceFromEtcd(now.Value)
		if err == nil {
			parsedNow = _parsedNow
		} else {
			log.Err(err).Msg("service looks invalid")
		}
	}
	if prev.Value != nil {
		_parsedPrev, err := validateServiceFromEtcd(prev.Value)
		if err == nil {
			parsedPrev = _parsedPrev
		} else {
			log.Err(err).Msg("could not marshal previous service state")
		}
	}

	if parsedNow == nil && parsedPrev == nil {
		// This happens when user created stuff manually badly
		l.Error().Msg("could not parse neither current nor previous version of this service: please check your service registry for invalid/inconsistent values ASAP")
		return nil, nil
	}

	parsedMetadata := []openapi.Metadata{}
	metadata := func() map[string]string {
		if parsedNow != nil {
			return parsedNow.Metadata
		}

		return parsedPrev.Metadata
	}()
	for key, val := range metadata {
		parsedMetadata = append(parsedMetadata, openapi.Metadata{Key: key, Value: val})
	}

	hadTarget := parsedPrev != nil && mapContainsKeys(parsedPrev.Metadata, e.options.targetKeys)
	hasTarget := parsedNow != nil && mapContainsKeys(parsedNow.Metadata, e.options.targetKeys)
	srv := parsedNow
	event := ""
	switch hasTarget {
	case false:
		if !hadTarget {
			log.Info().Msg("service doesn't have target keys and never had: skipping...")
			return nil, nil
		}
		log.Info().Msg("service doesn't have target keys anymore")
		event = "delete"
		srv = parsedPrev
	case true:
		if !hadTarget {
			log.Info().Msg("service now has target keys")
			event = "create"
		} else {
			if !mapValuesChanged(parsedNow.Metadata, parsedPrev.Metadata, e.options.targetKeys) {
				log.Info().Msg("no relevant changes found, skipping...")
				return nil, nil
			}
			event = "update"
		}
	}

	// if you're here, it means that there are indeed changes to be made.
	endpList, err := e.servreg.ListEndp(srv.NsName, srv.Name)
	if err != nil {
		log.Err(err).Msg("could not get list of endpoints, skipping...")
		return nil, err
	}

	events := map[string]*openapi.Event{}
	for _, endp := range endpList {
		key := opetcd.KeyFromNames(endp.NsName, endp.ServName, endp.Name)
		events[key.String()] = createOpenapiEvent(endp, srv, event)
	}

	return events, nil
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
