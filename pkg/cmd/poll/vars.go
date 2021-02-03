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

package poll

const (
	pollUse   string = "poll cloudmap|servicedirectory [flags]"
	pollShort string = "poll a service registry to discover changes"
	pollLong  string = `poll uses a polling mechanism to detect changes to a
service registry. This means that the CN-WAN Reader will perform http calls to
the service registry, parse the result and see the difference.

This method is implemented only for those service registries that do not
provide a better way to do this: include --help or -h to know which ones are
included under this command`
	pollExample string = "poll cloudmap --region us-west-2 --credentials path/to/credentials/file"
)
