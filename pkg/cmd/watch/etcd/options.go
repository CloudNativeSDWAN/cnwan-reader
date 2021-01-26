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

// Options contans data needed to connect to the etcd cluster correctly
type Options struct {
	// Endpoints is a list of hosts and ports where etcd nodes are running
	Endpoints []Endpoint `yaml:"endpoints,omitempty"`
	// Credentials to connect to the cluster, if authentication mode is enabled
	Credentials *Credentials `yaml:"credentials,omitempty"`
	// Prefix where the service registry objects are stored
	Prefix string `yaml:"prefix,omitempty"`

	// targetKeys is a list of metadata keys to look for.
	// This is not dervied from etcd's own flags, so we make it unexported.
	targetKeys []string
}

// Endpoint is a container with host and port of an etcd node
type Endpoint struct {
	// Host of the etcd node
	Host string `yaml:"host,omitempty"`
	// Port where the etcd node is listening from
	Port int32 `yaml:"port,omitempty"`
}

// Credentials is a container with username and password for authenticating
// to etcd
type Credentials struct {
	// Username to authenticate as
	Username string `yaml:"username,omitempty"`
	// Password for this username
	Password string `yaml:"password,omitempty"`
}
