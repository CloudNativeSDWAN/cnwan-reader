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

package services

import (
	"testing"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	. "github.com/stretchr/testify/assert"
)

func TestGetChanges(t *testing.T) {
	stored := map[string]*openapi.Service{
		"first": {
			Address:  "10.10.10.10",
			Port:     80,
			Metadata: []openapi.Metadata{{Key: "first-key", Value: "first-value"}},
			Name:     "first-name",
		},
		"second": {
			Address:  "11.11.11.11",
			Port:     8080,
			Metadata: []openapi.Metadata{{Key: "second-key", Value: "second-value"}},
			Name:     "second-name",
		},
		"third": {
			Address:  "12.12.12.12",
			Port:     80,
			Metadata: []openapi.Metadata{{Key: "third-key", Value: "third-value"}},
			Name:     "third-name",
		},
	}
	pulled := map[string]*openapi.Service{
		"first": {
			Address:  "10.10.10.10",
			Port:     80,
			Metadata: []openapi.Metadata{{Key: "first-key", Value: "first-value"}},
			Name:     "first-name",
		},
		"second": {
			Address:  "11.11.11.11",
			Port:     8080,
			Metadata: []openapi.Metadata{{Key: "second-key", Value: "second-value"}},
			Name:     "second-name",
		},
		"third": {
			Address:  "12.12.12.12",
			Port:     80,
			Metadata: []openapi.Metadata{{Key: "third-key", Value: "third-value"}},
			Name:     "third-name",
		},
	}

	// Case 1: nothing is changed
	res := getChanges(stored, pulled)
	Empty(t, res)

	// Case 2: something is new
	pulled["fourth"] = &openapi.Service{
		Address:  "13.13.13.13",
		Port:     9090,
		Metadata: []openapi.Metadata{{Key: "fourth-key", Value: "fourth-value"}},
		Name:     "fourth-name",
	}
	pulled["fifth"] = &openapi.Service{
		Address:  "14.14.14.14",
		Port:     9009,
		Metadata: []openapi.Metadata{{Key: "fifth-key", Value: "fifth-value"}},
		Name:     "fifth-name",
	}
	expectedRes := map[string]*openapi.Event{
		"fourth": {
			Event: "create",
			Service: openapi.Service{
				Address:  "13.13.13.13",
				Port:     9090,
				Metadata: []openapi.Metadata{{Key: "fourth-key", Value: "fourth-value"}},
				Name:     "fourth-name",
			},
		},
		"fifth": {
			Event: "create",
			Service: openapi.Service{
				Address:  "14.14.14.14",
				Port:     9009,
				Metadata: []openapi.Metadata{{Key: "fifth-key", Value: "fifth-value"}},
				Name:     "fifth-name",
			},
		},
	}
	res = getChanges(stored, pulled)
	Equal(t, expectedRes, res)

	// Case 3: something is changed
	// Remove these so pulled gets back to previous state
	delete(pulled, "fourth")
	delete(pulled, "fifth")

	pulled["first"].Metadata[0].Value = "first-changed-value"
	delete(pulled, "second")
	expectedRes = map[string]*openapi.Event{
		"second": {
			Event: "delete",
			Service: openapi.Service{
				Address:  "11.11.11.11",
				Port:     8080,
				Metadata: []openapi.Metadata{{Key: "second-key", Value: "second-value"}},
				Name:     "second-name",
			},
		},
		"first": {
			Event: "update",
			Service: openapi.Service{
				Address:  "10.10.10.10",
				Port:     80,
				Metadata: []openapi.Metadata{{Key: "first-key", Value: "first-changed-value"}},
				Name:     "first-name",
			},
		},
	}
	res = getChanges(stored, pulled)
	Equal(t, expectedRes, res)
}
