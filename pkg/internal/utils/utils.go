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

package utils

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/configuration"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// GetMetadataKeysFromCmdFlags returns the keys from --metadata-keys flag
func GetMetadataKeysFromCmdFlags(cmd *cobra.Command) ([]string, error) {
	keys := []string{}

	if cmd.Flags().Changed("metadata-keys") {
		keys, _ = cmd.Flags().GetStringSlice("metadata-keys")
	} else {
		if conf := configuration.GetConfigFile(); conf != nil && len(conf.MetadataKeys) > 0 {
			keys = conf.MetadataKeys
		}
	}

	switch l := len(keys); {
	case l == 0:
		return nil, fmt.Errorf("no metadata keys provided")
	case l > 1:
		log.Warn().Msg("multiple metadata keys are not supported yet, only the first one will be used")
		fallthrough
	default:
		return []string{keys[0]}, nil
	}
}

// MapContainsKeys returns true if the subject map contains target keys
func MapContainsKeys(subject map[string]string, targets []string) bool {
	foundKeys := 0
	for _, targetKey := range targets {
		if _, exists := subject[targetKey]; exists {
			foundKeys++
		}
	}

	return foundKeys == len(targets)
}

// GetAdaptorEndpointFromFlags gets the value of --adaptor-api or returns an
// error in case it is not valid.
func GetAdaptorEndpointFromFlags(cmd *cobra.Command) (string, error) {
	endp := "localhost:80/cnwan"

	if cmd.Flags().Changed("adaptor-api") {
		endp, _ = cmd.Flags().GetString("adaptor-api")
	} else {
		if conf := configuration.GetConfigFile(); conf != nil && len(conf.Adaptor) > 0 {
			endp = conf.Adaptor
		}
	}

	_endp := fmt.Sprintf("http://%s", endp)
	if _, err := url.ParseRequestURI(_endp); err != nil {
		return "", err
	}

	if strings.HasPrefix(endp, "localhost") {
		return SanitizeLocalhost(endp)
	}

	return endp, nil
}

// GetDebugModeFromFlags gets the value of --debug flag
func GetDebugModeFromFlags(cmd *cobra.Command) bool {
	if cmd.Flags().Changed("debug") {
		_debug, _ := cmd.Flags().GetBool("debug")
		return _debug
	}

	if conf := configuration.GetConfigFile(); conf != nil {
		return conf.DebugMode
	}

	return false
}

// SanitizeLocalhost changes localhost to host.docker.internal in case the
// project is running as a docker container.
//
// Of course, this function won't work with values like localhostS or
// localhost-example.com, but that is intentionally left un-implemented
// as that is a very corner case.
func SanitizeLocalhost(host string) (string, error) {
	_host := host
	if strings.HasPrefix(_host, "https://") {
		_host = _host[len("https://"):]
	}
	if strings.HasPrefix(_host, "http://") {
		_host = _host[len("http://"):]
	}

	_host = strings.Trim(_host, "/")

	if len(_host) == 0 {
		return "", fmt.Errorf("invalid host provided: %s", host)
	}

	if !strings.HasPrefix(_host, "localhost") {
		return _host, nil
	}

	if mode := os.Getenv("MODE"); mode != "docker" {
		return _host, nil
	}

	return strings.Replace(_host, "localhost", "host.docker.internal", 1), nil
}
