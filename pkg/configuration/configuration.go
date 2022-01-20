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

import (
	"io/ioutil"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	conf *Config
)

// ParseConfigurationFile parses the configuration file starting from the
// command
func ParseConfigurationFile(cmd *cobra.Command) (err error) {
	if conf != nil {
		return
	}

	if !cmd.Flags().Changed("conf") {
		return
	}

	filePath, _ := cmd.Flags().GetString("conf")

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	var _conf Config
	err = yaml.Unmarshal(yamlFile, &_conf)
	if err != nil {
		return
	}

	if len(_conf.MetadataKeys) > 0 {
		_conf.MetadataKeys = []string{_conf.MetadataKeys[0]}
	}

	conf = &_conf
	return
}

// GetConfigFile returns the configuration file parsed with
// ParseConfigurationFile.
// If the configuration file was not provided via --conf, then this returns
// nil.
func GetConfigFile() *Config {
	return conf
}
