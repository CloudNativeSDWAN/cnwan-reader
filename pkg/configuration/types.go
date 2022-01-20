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

package configuration

// Config contains the configuration of the program
type Config struct {
	// DebugMode specifies whether to log debug or not
	DebugMode bool `yaml:"debugMode,omitempty"`
	// Adaptor specifies the adaptor configuration
	Adaptor string `yaml:"adaptor,omitempty"`
	// MetadataKeys is the key to look for in a service's metadata
	MetadataKeys []string `yaml:"metadataKeys"`
	// ServiceRegistry settings about the service registry to use
	ServiceRegistry *ServiceRegistrySettings `yaml:"serviceRegistry"`
}

// ServiceRegistrySettings contains information
type ServiceRegistrySettings struct {
	// GCPServiceDirectory is the field with configuration about service
	// directory
	GCPServiceDirectory *ServiceDirectoryConfig `yaml:"gcpServiceDirectory,omitempty"`
	// AWSCloudMap contains configuration about AWS CloudMap
	AWSCloudMap *CloudMapConfig `yaml:"awsCloudMap,omitempty"`
}

// ServiceDirectoryConfig contains Service Directory configuration.
// Its fields are the same as the CLI flags, although the latter can override
// them.
type ServiceDirectoryConfig struct {
	// PollingInterval is the number of seconds between two consecutive polls
	PollingInterval int `yaml:"pollInterval,omitempty"`
	// ProjectID is the name of the Google Cloud project
	ProjectID string `yaml:"projectID"`
	// Region where to look for
	Region string `yaml:"region"`
	// ServiceAccountPath is the path of the service account JSON
	ServiceAccountPath string `yaml:"serviceAccountPath"`
}

// CloudMapConfig contans data need to connect to AWS Cloud Map correctly.
type CloudMapConfig struct {
	// Region where to look for
	Region string `yaml:"region,omitempty"`
	// CredentialsPath is the path where to find the AWS credentials.
	CredentialsPath string `yaml:"credentialsPath,omitempty"`
	// PollInterval is the number of seconds between two consecutive polls
	PollInterval int `yaml:"pollInterval,omitempty"`
}
