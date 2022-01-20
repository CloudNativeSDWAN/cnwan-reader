// Copyright © 2020 Cisco
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

package queue

import (
	"context"
	"testing"
	"time"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	assert "github.com/stretchr/testify/assert"
)

var (
	result chan int
)

type fakeHandler struct {
	t *testing.T
}

func (f *fakeHandler) Send(events []openapi.Event) error {
	// Emulate process
	time.Sleep(5 * time.Second)
	result <- len(events)
	return nil
}

func TestEnqueue(t *testing.T) {
	result = make(chan int)
	firstMap := map[string]*openapi.Event{
		"first": {
			Service: openapi.Service{
				Address:  "10.10.10.10",
				Port:     80,
				Metadata: []openapi.Metadata{{Key: "first-key", Value: "first-value"}},
				Name:     "first-name",
			},
		},
	}
	secondMap := map[string]*openapi.Event{
		"second": {
			Service: openapi.Service{
				Address:  "11.11.11.11",
				Port:     8080,
				Metadata: []openapi.Metadata{{Key: "second-key", Value: "second-value"}},
				Name:     "second-name",
			},
		},
		"third": {
			Service: openapi.Service{
				Address:  "12.12.12.12",
				Port:     80,
				Metadata: []openapi.Metadata{{Key: "third-key", Value: "third-value"}},
				Name:     "third-name",
			},
		},
		"fourth": {
			Service: openapi.Service{
				Address:  "13.13.13.13",
				Port:     8081,
				Metadata: []openapi.Metadata{{Key: "fourth-key", Value: "fourth-value"}},
				Name:     "fourth-name",
			},
		},
	}

	f := &fakeHandler{
		t: t,
	}

	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	q := New(ctx, f)
	go q.Enqueue(firstMap)
	time.Sleep(2 * time.Second)
	go q.Enqueue(secondMap)

	// Block here, for the first call
	firstCall := <-result
	// ... and the second
	secondCall := <-result

	if firstCall != len(firstMap) {
		assert.Fail(t, "first call had not 1 item but", firstCall)
	}
	if secondCall != len(secondMap) {
		assert.Fail(t, "second call had not 3 items but", secondCall)
	}
}
