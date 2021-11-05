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
	"testing"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/stretchr/testify/assert"
)

func TestGetServicesIDs(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		listServs func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error)

		expRes []string
		expErr error
	}{
		{
			listServs: func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error) {
				return nil, fmt.Errorf("any error")
			},
			expErr: fmt.Errorf("any error"),
		},
		{
			listServs: func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error) {
				return &servicediscovery.ListServicesOutput{
					Services: []*servicediscovery.ServiceSummary{},
				}, nil
			},
			expRes: []string{},
		},
		{
			listServs: func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error) {
				id := ""
				return &servicediscovery.ListServicesOutput{
					Services: []*servicediscovery.ServiceSummary{
						{
							Id: &id,
						},
					},
				}, nil
			},
			expRes: []string{},
		},
		{
			listServs: func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error) {
				id := "whatever"
				id1 := "whatever1"
				return &servicediscovery.ListServicesOutput{
					Services: []*servicediscovery.ServiceSummary{
						{Id: &id},
						{Id: &id1},
					},
				}, nil
			},
			expRes: []string{"whatever", "whatever1"},
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		cm := &awsCloudMap{
			sd: &fakeSD{
				_listServices: currCase.listServs,
			},
		}
		res, err := cm.getServicesIDs(context.Background())
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			failed(i)
		}
	}
}

func TestGetInstances(t *testing.T) {
	a := assert.New(t)
	instID1 := "inst-id"
	instID2 := "inst-id-2"
	ip4 := "10.10.10.10"
	ip6 := "2001:db8:a0b:12f0::1"
	port := int32(8989)
	cases := []struct {
		listInst func(ctx aws.Context, input *servicediscovery.ListInstancesInput, opts ...request.Option) (*servicediscovery.ListInstancesOutput, error)

		expRes []*openapi.Service
		expErr error
	}{
		{
			listInst: func(ctx aws.Context, input *servicediscovery.ListInstancesInput, opts ...request.Option) (*servicediscovery.ListInstancesOutput, error) {
				return nil, fmt.Errorf("any error")
			},
			expErr: fmt.Errorf("any error"),
		},
		{
			listInst: func(ctx aws.Context, input *servicediscovery.ListInstancesInput, opts ...request.Option) (*servicediscovery.ListInstancesOutput, error) {

				return &servicediscovery.ListInstancesOutput{
					Instances: []*servicediscovery.InstanceSummary{
						{
							Id: &instID1,
							Attributes: map[string]*string{
								"yes": &instID1,
								awsPortAttr: func() *string {
									p := "8989"
									return &p
								}(),
								awsIPv6Attr: &ip6,
							},
						},
						{
							Id: &instID2,
							Attributes: map[string]*string{
								"yes":       &instID2,
								awsIPv6Attr: &ip4,
							},
						},
					},
				}, nil
			},
			expRes: []*openapi.Service{
				{
					Name:     instID1,
					Address:  ip6,
					Port:     port,
					Metadata: []openapi.Metadata{{Key: "yes", Value: instID1}},
				},
				{
					Name:     instID2,
					Address:  ip4,
					Port:     int32(80),
					Metadata: []openapi.Metadata{{Key: "yes", Value: instID2}},
				},
			},
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		cm := &awsCloudMap{
			sd: &fakeSD{
				_listInstances: currCase.listInst,
			},
			opts: &options{
				keys: []string{"yes"},
			},
		}
		res, err := cm.getInstances(context.Background(), "whatever")
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			failed(i)
		}
	}
}

func TestParseInstance(t *testing.T) {
	cm := &awsCloudMap{
		opts: &options{
			keys: []string{"yes"},
		},
	}
	a := assert.New(t)
	srvID := "srv-id"
	instID := "inst-id"
	empty := ""
	ip4 := "10.10.10.10"
	ip6 := "2001:db8:a0b:12f0::1"
	port := int32(8989)
	attrs := map[string]*string{"yes": &instID}

	cases := []struct {
		inst *servicediscovery.InstanceSummary

		expRes *openapi.Service
		expErr error
	}{
		{
			inst:   &servicediscovery.InstanceSummary{},
			expErr: fmt.Errorf("found instance with no/empty ID"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {
				return &servicediscovery.InstanceSummary{Id: &empty}
			}(),
			expErr: fmt.Errorf("found instance with no/empty ID"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {
				return &servicediscovery.InstanceSummary{
					Id: &instID,
				}
			}(),
			expErr: fmt.Errorf("instance doesn't have any attribute"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {
				return &servicediscovery.InstanceSummary{
					Id: &instID,
					Attributes: map[string]*string{
						"ok": &empty,
					},
				}
			}(),
			expErr: fmt.Errorf("instance doesn't have required metadata keys"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {
				no := "no"
				return &servicediscovery.InstanceSummary{
					Id: &instID,
					Attributes: map[string]*string{
						"no": &no,
					},
				}
			}(),
			expErr: fmt.Errorf("instance doesn't have required metadata keys"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {

				return &servicediscovery.InstanceSummary{
					Id:         &instID,
					Attributes: attrs,
				}
			}(),
			expErr: fmt.Errorf("instance has no address"),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {

				return &servicediscovery.InstanceSummary{
					Id: &instID,
					Attributes: map[string]*string{
						"yes":       &srvID,
						awsIPv4Attr: &ip4,
					},
				}
			}(),
			expRes: func() *openapi.Service {
				return &openapi.Service{
					Name:     instID,
					Address:  ip4,
					Port:     awsDefaultInstancePort,
					Metadata: []openapi.Metadata{{Key: "yes", Value: srvID}},
				}
			}(),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {

				return &servicediscovery.InstanceSummary{
					Id: &instID,
					Attributes: map[string]*string{
						"yes":       &srvID,
						awsIPv6Attr: &ip6,
					},
				}
			}(),
			expRes: func() *openapi.Service {
				return &openapi.Service{
					Name:     instID,
					Address:  ip6,
					Port:     awsDefaultInstancePort,
					Metadata: []openapi.Metadata{{Key: "yes", Value: srvID}},
				}
			}(),
		},
		{
			inst: func() *servicediscovery.InstanceSummary {

				return &servicediscovery.InstanceSummary{
					Id: &instID,
					Attributes: map[string]*string{
						"yes": &srvID,
						awsPortAttr: func() *string {
							p := "8989"
							return &p
						}(),
						awsIPv6Attr: &ip6,
					},
				}
			}(),
			expRes: func() *openapi.Service {
				return &openapi.Service{
					Name:     instID,
					Address:  ip6,
					Port:     port,
					Metadata: []openapi.Metadata{{Key: "yes", Value: srvID}},
				}
			}(),
		},
	}

	failed := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res, err := cm.parseInstance(srvID, currCase.inst)
		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			failed(i)
		}
	}
}
