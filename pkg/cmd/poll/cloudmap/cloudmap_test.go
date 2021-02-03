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
