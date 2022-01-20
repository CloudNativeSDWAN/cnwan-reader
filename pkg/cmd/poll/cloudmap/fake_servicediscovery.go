// Copyright © 2021 Cisco
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

package cloudmap

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/aws/aws-sdk-go/service/servicediscovery/servicediscoveryiface"
)

type fakeSD struct {
	servicediscoveryiface.ServiceDiscoveryAPI

	_listServices  func(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error)
	_listInstances func(aws.Context, *servicediscovery.ListInstancesInput, ...request.Option) (*servicediscovery.ListInstancesOutput, error)
}

func (f *fakeSD) ListServicesWithContext(ctx aws.Context, input *servicediscovery.ListServicesInput, opts ...request.Option) (*servicediscovery.ListServicesOutput, error) {
	return f._listServices(ctx, input, opts...)
}

func (f *fakeSD) ListInstancesWithContext(ctx aws.Context, input *servicediscovery.ListInstancesInput, opts ...request.Option) (*servicediscovery.ListInstancesOutput, error) {
	return f._listInstances(ctx, input, opts...)
}
