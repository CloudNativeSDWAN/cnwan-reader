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
	opsr "github.com/CloudNativeSDWAN/cnwan-operator/pkg/servregistry"
	corev1 "k8s.io/api/core/v1"
)

type fakeSR struct {
	_getServ  func(nsName, servName string) (*opsr.Service, error)
	_listEndp func(nsName, servName string) ([]*opsr.Endpoint, error)
}

func (f *fakeSR) GetNs(name string) (*opsr.Namespace, error) {
	return nil, nil
}

func (f *fakeSR) ListNs() ([]*opsr.Namespace, error) {
	return nil, nil
}

func (f *fakeSR) CreateNs(ns *opsr.Namespace) (*opsr.Namespace, error) {
	return nil, nil
}

func (f *fakeSR) UpdateNs(ns *opsr.Namespace) (*opsr.Namespace, error) {
	return nil, nil
}

func (f *fakeSR) DeleteNs(name string) error {
	return nil
}

func (f *fakeSR) GetServ(nsName, servName string) (*opsr.Service, error) {
	return f._getServ(nsName, servName)
}

func (f *fakeSR) ListServ(nsName string) ([]*opsr.Service, error) {
	return nil, nil
}

func (f *fakeSR) CreateServ(serv *opsr.Service) (*opsr.Service, error) {
	return nil, nil
}

func (f *fakeSR) UpdateServ(serv *opsr.Service) (*opsr.Service, error) {
	return nil, nil
}

func (f *fakeSR) DeleteServ(nsName, servName string) error {
	return nil
}

func (f *fakeSR) GetEndp(nsName, servName, endpName string) (*opsr.Endpoint, error) {
	return nil, nil
}

func (f *fakeSR) ListEndp(nsName, servName string) ([]*opsr.Endpoint, error) {
	return f._listEndp(nsName, servName)
}

func (f *fakeSR) CreateEndp(endp *opsr.Endpoint) (*opsr.Endpoint, error) {
	return nil, nil
}

func (f *fakeSR) UpdateEndp(endp *opsr.Endpoint) (*opsr.Endpoint, error) {
	return nil, nil
}

func (f *fakeSR) DeleteEndp(nsName, servName, endpName string) error {
	return nil
}

func (f *fakeSR) ExtractData(ns *corev1.Namespace, serv *corev1.Service) (*opsr.Namespace, *opsr.Service, []*opsr.Endpoint, error) {
	return nil, nil, nil, nil
}
