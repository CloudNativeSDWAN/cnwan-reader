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
	"reflect"
	"sync"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
)

// Datastore holds services in their current state
type Datastore interface {
	// GetEvents receives the current services and runs a difference between
	// them and their previous state (the one already existing in memory).
	// It returns the differences in form of events.
	GetEvents(services map[string]*openapi.Service) map[string]*openapi.Event
}

type servicesDatastore struct {
	lock     sync.Mutex
	services map[string]*openapi.Service
}

// NewDatastore returns a new services datastore
func NewDatastore() Datastore {
	return &servicesDatastore{
		services: map[string]*openapi.Service{},
	}
}

// GetEvents receives the current services and runs a difference between
// them and their previous state (the one already existing in memory).
// It returns the differences in form of events.
func (m *servicesDatastore) GetEvents(currServices map[string]*openapi.Service) map[string]*openapi.Event {
	m.lock.Lock()
	defer m.lock.Unlock()

	//----------------------------------
	// Run difference
	//----------------------------------

	changes := getChanges(m.services, currServices)

	//----------------------------------
	// Update the services
	//----------------------------------

	for key, change := range changes {

		if change.Event == "delete" {
			delete(m.services, key)
		} else {
			m.services[key] = &change.Service
		}
	}

	return changes
}

func getChanges(storedState, currentState map[string]*openapi.Service) map[string]*openapi.Event {
	changes := map[string]*openapi.Event{}

	// Run the difference
	for currKey, currVal := range currentState {
		storedVal, exists := storedState[currKey]

		if !exists {
			// This is new
			changes[currKey] = &openapi.Event{
				Event:   "create",
				Service: *currVal,
			}
			continue
		}

		if !reflect.DeepEqual(storedVal, currVal) {
			// This is changed
			changes[currKey] = &openapi.Event{
				Event:   "update",
				Service: *currVal,
			}
		}

	}

	for storedKey, storedVal := range storedState {
		if _, exists := currentState[storedKey]; !exists {
			// This does not exist anymore
			changes[storedKey] = &openapi.Event{
				Event:   "delete",
				Service: *storedVal,
			}
		}
	}

	return changes
}
