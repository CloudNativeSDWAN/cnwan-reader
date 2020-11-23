// Copyright © 2020 Cisco
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

// Package cmd contains code that is executed by a given command from
// the CLI.
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

// Config contains the configuration of the program
type Config struct {
	// DebugMode specifies whether to log debug or not
	DebugMode bool `yaml:"debugMode,omitempty"`
	// Adaptor specifies the adaptor configuration
	Adaptor *AdaptorConfig `yaml:"adaptor,omitempty"`
	// MetadataKeys is the key to look for in a service's metadata
	MetadataKeys []string `yaml:"metadataKeys"`
	// ServiceRegistry settings about the service registry to use
	ServiceRegistry *ServiceRegistrySettings `yaml:"serviceRegistry"`
}

// AdaptorConfig contains configuration about the adaptor
type AdaptorConfig struct {
	// Host where the adaptor is running
	Host string `yaml:"host,omitempty"`
	// Port where the adaptor is listening from
	Port int32 `yaml:"port,omitempty"`
}

// ServiceRegistrySettings contains information
type ServiceRegistrySettings struct {
	// GCPServiceDirectory is the field with configuration about service
	// directory
	GCPServiceDirectory *ServiceDirectoryConfig `yaml:"gcpServiceDirectory,omitempty"`
}

// ServiceDirectoryConfig contains Service Directory configuration.
// Its fields are the same as the CLI flags, although the latter can override
// them.
type ServiceDirectoryConfig struct {
	// PollingInterval is number of seconds between two consecutive polls
	PollingInterval int `yaml:"pollInterval,omitempty"`
	// ProjectID is the name of the Google Cloud project
	ProjectID string `yaml:"projectID"`
	// Region where to look for
	Region string `yaml:"region"`
	// ServiceAccountPath is the path of the service account JSON
	ServiceAccountPath string `yaml:"serviceAccountPath"`
}

func parseConfigFile(filePath string) *Config {
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Fatal().Str("conf", filePath).Msg("file was not found")
			return nil // un-necessary, but just for clarity
		}

		logger.Fatal().Err(err).Str("conf", filePath).Msg("could not open this file")
		return nil
	}

	var conf Config
	if err := yaml.Unmarshal(yamlFile, &conf); err != nil {
		logger.Fatal().Err(err).Str("conf", filePath).Msg("error while unmarshaling config file")
		return nil
	}

	// Set up persistent flags first
	debugMode = conf.DebugMode

	switch keys := len(conf.MetadataKeys); {
	case keys == 0:
		logger.Fatal().Msg("no metadata keys have been provided: please provide at least one")
		return nil
	case keys > 1:
		// TODO: support this on next versions
		logger.Warn().Msg("watching multiple keys is not supported yet: only first one will be watched")
		fallthrough
	default:
		metadataKey = conf.MetadataKeys[0]
	}

	if conf.Adaptor != nil {
		host := conf.Adaptor.Host
		if len(host) == 0 {
			logger.Info().Str("default-host", "localhost").Msg("adaptor host not provided, using default...")
			host = "localhost"
		}
		port := conf.Adaptor.Port
		if port <= 0 {
			if port < 0 {
				logger.Error().Int32("provided", port).Int32("default", 80).Msg("invalid adaptor port provided, using default...")
			} else {
				logger.Info().Str("default-host", "localhost").Msg("adaptor host not provided, using default...")
			}
			port = 80
		}

		endpoint = sanitizeAdaptorEndpoint(fmt.Sprintf("%s:%d", host, port))
	}

	// TODO: check other service registries as well
	srConf := conf.ServiceRegistry
	if srConf != nil && srConf.GCPServiceDirectory != nil {
		parseServiceDirectoryConf(srConf.GCPServiceDirectory)
	}

	return &conf
}
