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

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func sanitizeLocalhost(host string, mode string) (string, error) {
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

	if strings.HasPrefix(_host, "localhost") && mode == "docker" {
		_host = fmt.Sprintf("%s%s", "host.docker.internal", _host[len("localhost"):])
	}

	return _host, nil
}

func parseEndpointsFromFlags(endpoints []string) (parsed []Endpoint) {
	mode := os.Getenv("MODE")
	dups := map[string]bool{}

	for _, endp := range endpoints {
		if _, exists := dups[endp]; exists {
			log.Warn().Str("endpoint", endp).Msg("found duplicate endpoint: skipping...")
			continue
		}

		split := strings.Split(endp, ":")
		host := defaultHost
		port := defaultPort

		switch l := len(split); {
		case l > 2:
			log.Error().Str("endpoint", endp).Msg("skipping invalid endpoint")
			continue
		case l == 1:
			host = split[0]
		case l == 2:
			host = split[0]
			if len(split[1]) > 0 {
				_port, err := strconv.ParseInt(split[1], 10, 32)
				if err != nil {
					log.Error().Str("endpoint", endp).Msg("could not parse port: skipping...")
					continue
				}
				port = int32(_port)
			}
		}

		host, err := sanitizeLocalhost(host, mode)
		if err != nil {
			log.Err(err).Msg("error while parsing endpoint, skipping...")
			continue
		}

		parsed = append(parsed, Endpoint{Host: host, Port: port})
		dups[endp] = true
	}

	return parsed
}

func parseFlags(cmd *cobra.Command) (*Options, error) {
	opts := &Options{}

	endpoints, _ := cmd.Flags().GetStringSlice("endpoints")
	opts.Endpoints = parseEndpointsFromFlags(endpoints)

	keys := []string{}
	_keys, _ := cmd.Flags().GetStringSlice("metadata-keys")
	switch l := len(_keys); {
	case l == 0:
		return nil, fmt.Errorf("no metadata keys provided")
	case l > 1:
		log.Warn().Msg("multiple metadata keys are not supported yet, only the first one will be used")
		fallthrough
	default:
		keys = append(keys, _keys[0])
	}

	opts.targetKeys = keys

	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	if len(username) > 0 && len(password) > 0 {
		opts.Credentials = &Credentials{Username: username, Password: password}
	} else {
		if len(username) > 0 && len(password) == 0 {
			return nil, fmt.Errorf("username set but no password provided")
		}
		if len(username) == 0 && len(password) > 0 {
			return nil, fmt.Errorf("password set but no username provided")
		}
	}

	prefix, _ := cmd.Flags().GetString("prefix")
	opts.Prefix = parsePrefix(prefix)

	return opts, nil
}

func parsePrefix(prefix string) string {
	if len(prefix) == 0 || prefix == "/" {
		return "/"
	}

	// Remove all slashes to prevent having values like //key////
	pref := strings.Trim(prefix, "/")
	return fmt.Sprintf("/%s/", pref)
}
