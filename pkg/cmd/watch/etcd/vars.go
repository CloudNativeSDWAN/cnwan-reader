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

const (
	etcdUse   string = "etcd [flags]"
	etcdShort string = "watch for changes in etcd"
	etcdLong  string = `etcd command connects to the etcd cluster and watches
for changes according to the options that you specified as flags.

--endpoints is a list of node addresses in the form of host:port where etcd is running

--username and --password must be both not empty if your cluster has authentication
mode enabled. Otherwise, you can leave both empty if you are allowing guest users to
use your cluster (not recommended!). 

--prefix is a string that will be placed before each query and specifies that your service
registry objects, i.e. Namespaces, Services and Endpoints, all have keys that start with
these values. If you have authentication mode enabled, make sure your user has enough
permissions for the cnwan-reader to do its job: it must have read access to this prefix.`
	etcdExample string = "etcd --endpoints localhost:2379 --username user --password pass"

	defaultPort int32  = 2379
	defaultHost string = "localhost"
)
