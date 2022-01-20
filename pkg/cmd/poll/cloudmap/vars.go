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

const (
	cmdUse   string = "cloudmap --region <region> [--credentials-path <credentials-path>]"
	cmdShort string = "connect to Cloud Map to get registered services"
	cmdLong  string = `cloudmap connects to AWS CloudMap and
observes changes to services published in it, i.e. metadata, addresses and
ports.
	
For this to work, a valid region must be provided with --region and the
aws credentials must be properly set.

Unless a different credentials path is defined with --credentials-path,
$HOME/.aws/credentials on Linux/Unix and %USERPROFILE%\.aws\credentials on
Windows will be used instead. Alternatively, credentials path can be set
with environment variables. For a complete list of alternatives, please
refer to AWS Session documentation, but, to keep things simple, we suggest you
use the default one.`
	cmdExample string = "cloudmap --region us-west-2 --credentials path/to/credentials/file"
)
