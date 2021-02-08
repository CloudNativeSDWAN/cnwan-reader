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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapContainsKeys(t *testing.T) {
	a := assert.New(t)
	m := map[string]string{
		"key1": "val1",
		"key2": "val2",
		"key3": "val3",
	}
	cases := []struct {
		targets []string
		expRes  bool
	}{
		{
			targets: []string{"key1"},
			expRes:  true,
		},
		{
			targets: []string{"key1", "key2", "key4"},
			expRes:  false,
		},
		{
			targets: []string{"key1", "key2", "key3", "key4"},
			expRes:  false,
		},
	}

	fail := func(i int) {
		a.FailNow("case failed", fmt.Sprintf("case %d", i))
	}
	for i, currCase := range cases {
		res := MapContainsKeys(m, currCase.targets)
		if !a.Equal(currCase.expRes, res) {
			fail(i)
		}
	}
}

func TestSanitizeLocalhost(t *testing.T) {
	a := assert.New(t)

	cases := []struct {
		host   string
		mode   string
		expRes string
		expErr error
	}{
		{
			host:   "http://",
			expErr: fmt.Errorf("invalid host provided: %s", "http://"),
		},
		{
			host:   "https://",
			expErr: fmt.Errorf("invalid host provided: %s", "https://"),
		},
		{
			host:   "non-localhost",
			expRes: "non-localhost",
		},
		{
			host:   "localhost/whatever",
			expRes: "localhost/whatever",
		},
		{
			host:   "localhost/whatever",
			mode:   "docker",
			expRes: "host.docker.internal/whatever",
		},
		{
			host:   "localhost/whatever",
			mode:   "whatever",
			expRes: "localhost/whatever",
		},
	}

	for i, currCase := range cases {
		os.Clearenv()
		if len(currCase.mode) > 0 {
			os.Setenv("MODE", currCase.mode)
		}
		res, err := SanitizeLocalhost(currCase.host)

		if !a.Equal(currCase.expRes, res) || !a.Equal(currCase.expErr, err) {
			a.FailNow(fmt.Sprintf("case %d failed", i))
		}
	}
}
