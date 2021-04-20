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

package kubernetes

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAreMetadataEqual(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		old    []openapi.Metadata
		curr   []openapi.Metadata
		expRes bool
	}{
		{
			old:    []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}},
			curr:   []openapi.Metadata{{Key: "one", Value: "one-val"}},
			expRes: false,
		},
		{
			old:    []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}},
			curr:   []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}, {Key: "three", Value: "three-val"}},
			expRes: false,
		},
		{
			old:    []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}},
			curr:   []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val-changed"}},
			expRes: false,
		},
		{
			old:    []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}},
			curr:   []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "three", Value: "three-val"}},
			expRes: false,
		},
		{
			old:    []openapi.Metadata{{Key: "one", Value: "one-val"}, {Key: "two", Value: "two-val"}},
			curr:   []openapi.Metadata{{Key: "two", Value: "two-val"}, {Key: "one", Value: "one-val"}},
			expRes: true,
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		if areMetadataEqual(currCase.old, currCase.curr) != currCase.expRes {
			failed(i)
		}
	}
}

func TestGetFromK8sService(t *testing.T) {
	a := assert.New(t)

	keys := []string{"target-key"}
	val := "val"
	ips := []string{"10.10.10.10", "10.10.10.11"}
	ports := []int32{80, 8080}
	nsName, servName := "ns", "serv"

	cases := []struct {
		serv   *corev1.Service
		expRes []*openapi.Service
		expErr error
	}{
		{
			serv: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: servName, Namespace: nsName},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
				},
			},
			expErr: fmt.Errorf("service is not of type LoadBalancer"),
		},
		{
			serv: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: servName, Namespace: nsName},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{},
			},
			expErr: fmt.Errorf("service has no LoadBalancer IPs"),
		},
		{
			serv: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: servName, Namespace: nsName},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{IP: ips[0]}, {IP: ips[1]}},
					},
				},
			},
			expErr: fmt.Errorf("service does not have required annotations"),
		},
		{
			serv: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:        servName,
					Namespace:   nsName,
					Annotations: map[string]string{keys[0]: val},
				},
				Spec: corev1.ServiceSpec{
					Type:  corev1.ServiceTypeLoadBalancer,
					Ports: []corev1.ServicePort{{Port: ports[0]}, {Port: ports[1]}},
				},
				Status: corev1.ServiceStatus{
					LoadBalancer: corev1.LoadBalancerStatus{
						Ingress: []corev1.LoadBalancerIngress{{IP: ips[0]}, {IP: ips[1]}},
					},
				},
			},
			expRes: []*openapi.Service{
				{
					Name: func() string {
						h := sha256.New()
						h.Write([]byte(fmt.Sprintf("%s:%d", ips[0], ports[0])))
						name := fmt.Sprintf("%s/%s-%s", nsName, servName, hex.EncodeToString(h.Sum(nil))[:10])
						return name
					}(),
					Address:  ips[0],
					Port:     ports[0],
					Metadata: []openapi.Metadata{{Key: keys[0], Value: val}},
				},
				{
					Name: func() string {
						h := sha256.New()
						h.Write([]byte(fmt.Sprintf("%s:%d", ips[0], ports[1])))
						name := fmt.Sprintf("%s/%s-%s", nsName, servName, hex.EncodeToString(h.Sum(nil))[:10])
						return name
					}(),
					Address:  ips[0],
					Port:     ports[1],
					Metadata: []openapi.Metadata{{Key: keys[0], Value: val}},
				},
				{
					Name: func() string {
						h := sha256.New()
						h.Write([]byte(fmt.Sprintf("%s:%d", ips[1], ports[0])))
						name := fmt.Sprintf("%s/%s-%s", nsName, servName, hex.EncodeToString(h.Sum(nil))[:10])
						return name
					}(),
					Address:  ips[1],
					Port:     ports[0],
					Metadata: []openapi.Metadata{{Key: keys[0], Value: val}},
				},
				{
					Name: func() string {
						h := sha256.New()
						h.Write([]byte(fmt.Sprintf("%s:%d", ips[1], ports[1])))
						name := fmt.Sprintf("%s/%s-%s", nsName, servName, hex.EncodeToString(h.Sum(nil))[:10])
						return name
					}(),
					Address:  ips[1],
					Port:     ports[1],
					Metadata: []openapi.Metadata{{Key: keys[0], Value: val}},
				},
			},
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res, err := getDataFromK8sService(currCase.serv, keys)
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			failed(i)
		}
	}
}

func TestGetServChanges(t *testing.T) {
	a := assert.New(t)
	old := []*openapi.Service{
		{
			Name:     "serv-1",
			Address:  "10.10.10.10",
			Port:     8181,
			Metadata: []openapi.Metadata{{Key: "s1-key-1", Value: "val-1"}, {Key: "s1-key-2", Value: "val-2"}},
		},
		{
			Name:     "serv-2",
			Address:  "20.20.20.20",
			Port:     8282,
			Metadata: []openapi.Metadata{{Key: "s2-key-1", Value: "val-1"}, {Key: "s2-key-2", Value: "val-2"}},
		},
		{
			Name:     "serv-3",
			Address:  "30.30.30.30",
			Port:     8383,
			Metadata: []openapi.Metadata{{Key: "s3-key-1", Value: "val-1"}, {Key: "s3-key-2", Value: "val-2"}},
		},
	}

	cases := []struct {
		old    []*openapi.Service
		curr   []*openapi.Service
		expRes []*openapi.Event
	}{
		{
			old: old,
			curr: func() []*openapi.Service {
				cpy := make([]*openapi.Service, len(old))
				copy(cpy, old)

				cpy[1] = &openapi.Service{
					Name:     "serv-2",
					Address:  "20.20.20.20",
					Port:     8282,
					Metadata: []openapi.Metadata{{Key: "s2-key-1", Value: "changed"}, {Key: "s2-key-2", Value: "val-2"}},
				}
				cpy[2] = &openapi.Service{
					Name:     "serv-4",
					Address:  "40.40.40.40",
					Port:     8484,
					Metadata: []openapi.Metadata{{Key: "s4-key-1", Value: "s4-val-1"}, {Key: "s4-key-2", Value: "s4-val-2"}},
				}

				return cpy
			}(),
			expRes: func() []*openapi.Event {
				evs := []*openapi.Event{
					{
						Event:   "delete",
						Service: *old[2],
					},
					{
						Event: "update",
						Service: openapi.Service{
							Name:     old[1].Name,
							Address:  old[1].Address,
							Port:     old[1].Port,
							Metadata: []openapi.Metadata{{Key: "s2-key-1", Value: "changed"}, {Key: "s2-key-2", Value: "val-2"}},
						},
					},
					{
						Event: "create",
						Service: openapi.Service{
							Name:     "serv-4",
							Address:  "40.40.40.40",
							Port:     8484,
							Metadata: []openapi.Metadata{{Key: "s4-key-1", Value: "s4-val-1"}, {Key: "s4-key-2", Value: "s4-val-2"}},
						},
					},
				}

				return evs
			}(),
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		if !a.ElementsMatch(currCase.expRes, getServChanges(currCase.old, currCase.curr)) {
			failed(i)
		}
	}
}
