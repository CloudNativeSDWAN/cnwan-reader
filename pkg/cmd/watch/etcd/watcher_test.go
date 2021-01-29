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

package etcd

import (
	"context"
	"fmt"
	"testing"

	opsr "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry"
	opetcd "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry/etcd"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gopkg.in/yaml.v2"
)

func TestGetCurrentState(t *testing.T) {
	a := assert.New(t)
	okSrv := opsr.Service{
		Name:   "should-stay",
		NsName: "whatever",
		Metadata: map[string]string{
			"stay": "yes",
		},
	}
	okSrvKey := opetcd.KeyFromNames(okSrv.NsName, okSrv.Name)
	okSrvBytes, _ := yaml.Marshal(okSrv)

	koSrv := opsr.Service{
		Name:   "should-not-stay",
		NsName: "whatever",
		Metadata: map[string]string{
			"not-stay": "yes",
		},
	}
	koSrvKey := opetcd.KeyFromNames(koSrv.NsName, koSrv.Name)
	koSrvBytes, _ := yaml.Marshal(koSrv)

	okEp := opsr.Endpoint{
		Name:     "should-stay",
		ServName: "should-stay",
		NsName:   "whatever",
		Metadata: map[string]string{
			"whatever": "whatever",
		},
		Address: "10.10.10.10",
		Port:    8080,
	}
	okEpKey := opetcd.KeyFromNames(okEp.NsName, okEp.ServName, okEp.Name)
	okEpBytes, _ := yaml.Marshal(okEp)

	koEp := opsr.Endpoint{
		Metadata: map[string]string{
			"whatever": "whatever",
		},
		Address: "10.10.10.10",
		Port:    8080,
	}
	koEpKey := opetcd.KeyFromNames(okSrv.NsName, okSrv.Name, "empty")
	koEpBytes, _ := yaml.Marshal(koEp)
	invalidSrvKey := opetcd.KeyFromNames("ns", "srv")
	invalidEpKey := opetcd.KeyFromNames("ns", "srv", "ep")

	koEp1 := opsr.Endpoint{
		Name:     "should-not-stay",
		ServName: "should-not-stay",
		NsName:   "whatever",
		Metadata: map[string]string{
			"whatever": "whatever",
		},
		Address: "10.10.10.10",
		Port:    8080,
	}
	koEp1Key := opetcd.KeyFromNames(koEp1.NsName, koEp1.ServName, koEp1.Name)
	koEp1Bytes, _ := yaml.Marshal(koEp1)

	cases := []struct {
		event   string
		options *Options
		expRes  map[string]*openapi.Event
		get     func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
		expErr  error
	}{
		{
			event: "create",
			get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return nil, fmt.Errorf("any error")
			},
			expErr: fmt.Errorf("any error"),
		},
		{
			event: "event-type",
			get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{Key: []byte(okSrvKey.String()), Value: okSrvBytes},
						{Key: []byte(koSrvKey.String()), Value: koSrvBytes},
						{Key: []byte(okEpKey.String()), Value: okEpBytes},
						{Key: []byte(koEpKey.String()), Value: koEpBytes},
						{Key: []byte(koEp1Key.String()), Value: koEp1Bytes},
						{Key: []byte(invalidSrvKey.String()), Value: []byte("invalid")},
						{Key: []byte(invalidEpKey.String()), Value: []byte("invalid")},
					},
				}, nil
			},
			options: &Options{targetKeys: []string{"should-stay"}},
			expRes:  map[string]*openapi.Event{},
		},
		{
			event: "event-type",
			get: func(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
				return &clientv3.GetResponse{
					Kvs: []*mvccpb.KeyValue{
						{Key: []byte(okSrvKey.String()), Value: okSrvBytes},
						{Key: []byte(koSrvKey.String()), Value: koSrvBytes},
						{Key: []byte(okEpKey.String()), Value: okEpBytes},
						{Key: []byte(koEpKey.String()), Value: koEpBytes},
						{Key: []byte(koEp1Key.String()), Value: koEp1Bytes},
					},
				}, nil
			},
			options: &Options{targetKeys: []string{"stay"}},
			expRes: map[string]*openapi.Event{
				fmt.Sprintf("%s:%d", okEp.Address, okEp.Port): {
					Event: "event-type",
					Service: openapi.Service{
						Name:     okEp.Name,
						Address:  okEp.Address,
						Port:     okEp.Port,
						Metadata: []openapi.Metadata{{Key: "stay", Value: "yes"}},
					},
				},
			},
		},
	}
	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		e := &etcdWatcher{
			kv: &fakeKV{
				_get: currCase.get,
			},
			options: currCase.options,
		}
		res, err := e.getCurrentState(context.Background(), currCase.event)

		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			fail(i)
		}
	}
}

func TestParseEndpointAndCreateEvent(t *testing.T) {
	a := assert.New(t)
	ns := &opsr.Namespace{
		Name:     "ns",
		Metadata: map[string]string{"whatever": "whatever"},
	}
	srv := &opsr.Service{
		Name:     "srv",
		NsName:   ns.Name,
		Metadata: map[string]string{"yes": "yes"},
	}
	okEndp := &opsr.Endpoint{
		Name:     "ok-endp",
		ServName: srv.Name,
		NsName:   ns.Name,
		Address:  "10.10.10.10",
		Port:     9394,
		Metadata: map[string]string{"whatever": "whatever"},
	}
	okEndpKey, _ := opetcd.KeyFromServiceRegistryObject(okEndp)
	okEndpVal, _ := yaml.Marshal(okEndp)

	cases := []struct {
		kv        *mvccpb.KeyValue
		eventName string
		getServ   func(nsName, servName string) (*opsr.Service, error)

		expRes *openapi.Event
		expErr error
	}{
		{
			kv: &mvccpb.KeyValue{
				Key: []byte(opetcd.KeyFromNames("ns", "srv", "endp").String()),
				Value: func() []byte {
					endp := opsr.Endpoint{}
					ep, _ := yaml.Marshal(endp)
					return ep
				}(),
			},
			expErr: opsr.ErrNsNameNotProvided,
		},
		{
			kv: &mvccpb.KeyValue{
				Key:   []byte(okEndpKey.String()),
				Value: okEndpVal,
			},
			getServ: func(nsName, servName string) (*opsr.Service, error) {
				return nil, opsr.ErrNotFound
			},
			expErr: opsr.ErrNotFound,
		},
		{
			kv: &mvccpb.KeyValue{
				Key:   []byte(okEndpKey.String()),
				Value: okEndpVal,
			},
			getServ: func(nsName, servName string) (*opsr.Service, error) {
				return &opsr.Service{
					Name:     "name",
					NsName:   "nsname",
					Metadata: map[string]string{"no": "no"},
				}, nil
			},
		},
		{
			kv: &mvccpb.KeyValue{
				Key:   []byte(okEndp.Name),
				Value: okEndpVal,
			},
			eventName: "just-to-see-if-its-this",
			getServ: func(nsName, servName string) (*opsr.Service, error) {
				return srv, nil
			},
			expRes: &openapi.Event{
				Event: "just-to-see-if-its-this",
				Service: openapi.Service{
					Name:     okEndp.Name,
					Address:  okEndp.Address,
					Port:     okEndp.Port,
					Metadata: []openapi.Metadata{{Key: "yes", Value: "yes"}},
				},
			},
		},
	}

	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		e := &etcdWatcher{
			options: &Options{
				targetKeys: []string{"yes"},
			},
			servreg: &fakeSR{
				_getServ: currCase.getServ,
			},
		}

		res, err := e.parseEndpointAndCreateEvent(currCase.kv, currCase.eventName)
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			fail(i)
		}
	}
}
