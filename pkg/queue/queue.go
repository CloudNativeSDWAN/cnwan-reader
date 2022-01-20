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
	"sync"

	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/openapi"
	"github.com/CloudNativeSDWAN/cnwan-reader/pkg/services"
	"github.com/rs/zerolog/log"
)

// Queue contains data that will be sent to a handler
type Queue interface {
	// Enqueue intructs the queue that a new data must be sent on next request
	Enqueue(events map[string]*openapi.Event)
}

type senderWorkQueue struct {
	mainCtx      context.Context
	lock         sync.Mutex
	wakeUp       chan int
	queue        map[string]*openapi.Event
	servsHandler services.Handler
}

// New returns a Queue that receives data and sends it in bulk whenever
// possible
func New(ctx context.Context, servsHandler services.Handler) Queue {
	queue := &senderWorkQueue{
		mainCtx:      ctx,
		wakeUp:       make(chan int),
		queue:        map[string]*openapi.Event{},
		servsHandler: servsHandler,
	}

	go queue.work()

	return queue
}

// Enqueue intructs the queue that new data must be sent on next request
func (s *senderWorkQueue) Enqueue(events map[string]*openapi.Event) {
	wake := func() bool {
		s.lock.Lock()
		defer s.lock.Unlock()

		shouldWakeUp := true
		if len(s.queue) > 0 {
			// There was already something in the queue. It means that the
			// worker is already awake, no need to do it again.
			shouldWakeUp = false
		}

		for key, event := range events {
			s.queue[key] = event
		}

		return shouldWakeUp
	}()

	if wake {
		// Wake up the consumer
		// 0 is a dumb value
		s.wakeUp <- 0
	}
}

func (s *senderWorkQueue) work() {
	l := log.With().Str("func", "queue.senderWorkQueue.work").Logger()

	for {
		select {
		case <-s.wakeUp:
			l.Debug().Msg("worker woke up")
			// I have been woken up. This means there's work to do
			s.sendData()
		case <-s.mainCtx.Done():
			l.Info().Msg("stop requested")
			return
		}
	}
}

func (s *senderWorkQueue) sendData() {
	l := log.With().Str("func", "queue.senderWorkQueue.sendData").Logger()

	data := func() []openapi.Event {
		s.lock.Lock()
		defer s.lock.Unlock()
		events := make([]openapi.Event, 0, len(s.queue))

		// We copy the queue to an array so that we can directly send it,
		// this way we release the lock immediately, so other components
		// can enqueue new data while we're busy sending.
		for _, event := range s.queue {
			events = append(events, *event)
		}

		// Empty the queue, so we don't resend these values again
		s.queue = map[string]*openapi.Event{}

		return events
	}()

	l = l.With().Int("length", len(data)).Logger()
	l.Info().Msg("sending data...")

	if err := s.servsHandler.Send((data)); err != nil {
		// The error is logged from the service handler
		return
	}

	l.Info().Msg("events sent successfully")
}
